# Telemetry Server
Telemetry server for Georgia Tech Solar Racing

Handles listening to TCP data port, interfacing with InfluxDB for storage, and serving the strategy API.

## Prerequisites

* Windows 10 Pro/Linux/OS X
* [Docker](https://docs.docker.com/install/)
* [Docker-compose](https://docs.docker.com/compose/install/)

Note that modern Docker requires Windows 10 Pro. If you're on Windows 10 Basic or older version sof Windows, you're going to have to install a virtual machine or try to get it working with older versions of docker engine (not tested)

## Running
```
docker-compose up -d
```
Will automatically pull and build all relevant docker containers, and initialize them as specified in docker-compose.yml

## Exposed Ports

These ports are forwarded onto the docker host machine
Note that in production, only port 6001 and 80 are exposed to the internet. 

We rely on nginx as a reverse proxy to proxy traffic to the relevant local sockets

| Container | Port | Notes                                |
| --------- | ---- | -----------------------------------  |
| grafana   | 3000 | forwarded from grafana.solarracing.me|
| server    | 8888 | API entrypoint for solarracing.me api calls|
| server    | 6001 | TCP port exposed to internet for car to send data |
| nginx     | 80   | HTTP port exposed to internet |
| nginx     | 443  | HTTPS port exposed to internet (TODO once we get cert)|
| jenkins   | 8080 | Only accessible via tunnelling. See connecting to Jenkins below |

## Jenkins
We run a Jenkins Server in production that automatically manages deployments and runs tests before merges into master and pushes to branches.

The Jenkins *container* actually has full access to the Docker daemon on production. As a precaution against user error (but not necessarily making the system more secure), the jenkins user does not have direct root access to the server. Therefore, we have some pre-defined scripts `change-socket.docker` and `copy.docker` that run privledged to grant the dockerd unix socket and copy to /opt/telemetry-server on production. Note that since Jenkins has access to the Docker daemon, which runs as root, it should be treated as having root access.

Installing packages, running unit tests, and linting all occur in Docker containers as specified in the Jenkins pipeline. Actually copying the files occurs outside of the Docker containers

Also note that due to the way the go docker container works, we copy our go files into $GOPATH before we run our tests in docker. It is probably worth looking at fixing/avoiding this step to clean up the Jenkins code in the future.
 
## Connecting to Jenkins

This Jenkins server is only accessible by reverse tunneling into production.First you *must* connect to the Georgia Tech VPN (ssh traffic is only accessible within the Georgia Tech network, so eduroam/gtwifi won't work). Then, proxy-jenkins.sh will tunnel traffic on port 8080 to localhost:8088. *Note the final digit is different.*

Run

```
./proxy-jenkins.sh PRISM_ID
```
And then navigate to your web browser to localhost:8088 to access Jenkins jobs. We may in the future expose Jenkins to Georgia Tech IPs once we get an SSL cert and we deem it safe enough

## Architecture

## Data Protocol

CAN data is first parsed by the telemetry subsystem on the car. The microcontroller connected to the Cellular LTE modem will convert each CAN frame into a basic 15-byte format, and forward it via TCP to port 6001 on the server.

| Bytes 0-3 | Bytes 4-5 | Bytes 6-7 | Bytes 8-15 |
|  ---      |  ---      |    ---    |   ---      |
|  'GTSR'   | Unused    | CAN_ID    | Payload    |

All data is transmitted in little-endian format. Can Frames that do not use the whole 8 byte payload must still transmit the fixed 8 byte size. Perhaps the unusued two bytes may be useful for a size specifier or a checksum in this instance.

## Handling connections

Each inbound TCP connection is assigned a `ConnectionHandler` struct in `listener.go`. We spin a different goroutine for every `ConnectionHandler` that each has a `packetParser` state machine that is responsible for processing the above data format.

## Parsing Data

Our CAN configurations are stored in `configs/can_config.json`. This JSON file stores the mappings between each metric and an offset/datatype for an appropriate CAN_ID. The parser state machines validates the data parsed and sees if the incoming CAN_ID contains a valid metric. If it does, extracts the approprate data, and forwards it to a global DatapointPublisher channel

## Storage

All published datapoints are inserted into our influxdb database.

## Grafana

Grafana has built-in support for InfluxDB, and queries can be made via the Grafana dashboard. Note that a Datasource must be added that connects to http://influxdb:8086.

## Computational Clients

In addition to metrics reported by the car. We provide functionality for small programs to compute new metrics formed from the transformation of 1 or more other metrics. This is accomplished by subscribing to the datapoint publisher and fanning out a subset of relevant metric names to a channel to each computation client, which updates each computation client one data point at a time. Then, each computational client has the ability to write new metrics back into the global datapoint publisher channel, which can then be stored back into the storage layer.

## RF Listener

We interface with our RF subsystem by relaying our RF data to the tcp input of a server. The intention is that we can run all relevant parts of our server locallying on a laptop while trailering the car, and relay the RF data to localhost port 6001. We also provide the capability, if an internet connection is available, to relay the data to the server in production. This is primarily done via the Raspberry Pi in shop for debugging purposes.
