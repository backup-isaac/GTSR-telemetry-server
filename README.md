# Telemetry Server
Telemetry server for Georgia Tech Solar Racing

Handles listening to TCP data port, interfacing with InfluxDB for storage, and serving the strategy API.

## Prerequisites

* Windows 10 Home/Pro/Linux/OS X
* [Docker](https://docs.docker.com/install/)
* [Docker-compose](https://docs.docker.com/compose/install/)

Note that modern Docker requires Windows 10 Pro. If you're on Windows 10 Basic or older version sof Windows, you're going to have to install a virtual machine or try to get it working with older versions of docker engine (not tested)

## Initial Setup
Make sure to install above. Also, if you are running OSx, and would like to run the RF relay to allow you to connect to the RF antenna, 
you will also need to install go.

First you should build and update all containers

```
docker-compose build --pull
```

Then you should initialize all contianers

```
docker-compose up -d
```

If you are using Windows 10 Home, follow these steps, else continue past this section and go to "Then, once all the containers are initialized..."

Open "Oracle VM VirtualBox Manager" and open the "Settings" tab. Navigate to the "Network" tab and click on "Advanced" then "Port Forwarding".

There should already be a port here. Add ports corresponding to the "Exposed Ports" in this README.md file. When filling out the ports, use the same Host IP as the inital port and use the Port in the Exposed Port section for both the "Host Port" and "Guest Port". Click "Ok" till you get back to the main screen, and you can close the application.

Then, once all the containers are initialized, go to your web browser and navigate to http://grafana.localhost/ use admin/admin as your credentials.

Next, add a data source. The URL will be http://influxdb:8086 and the database name will be `telemetry` with no username or password. Name it what you'd like.

Now run

```
docker exec -it server go run generator/data_generator.go
```

This will create test data, to verify your pipeline is properly running, navigate to grafana, add a new dashboard, add a new panel, and select `Test` as your metric name.

If you see data being generated, then you have configured your development environment.


## Running
```
docker-compose up -d
```
Will automatically pull and build all relevant docker containers, and initialize them as specified in docker-compose.yml

## Exposed Ports

These ports are forwarded onto the docker host machine
Note that in production, only port 6001, 80, and 443 are exposed to the internet. 

We rely on nginx as a reverse proxy to proxy traffic to the relevant local sockets

| Container | Port | Notes                                |
| --------- | ---- | -----------------------------------  |
| grafana   | 3000 | forwarded from grafana.solarracing.me|
| server    | 8888 | API entrypoint for solarracing.me api calls|
| server    | 6001 | TCP port exposed to internet for car to send data |
| nginx     | 80   | HTTP port exposed to internet |
| nginx     | 443  | HTTPS port exposed to internet |
| jenkins   | 8080 | Accessible at jenkins.solarracing.me |

Note that the server telemetry API is available on HTTP and HTTPS, but grafana and jenkins can only be accessed via HTTPS

## Jenkins
We run a Jenkins Server in production that automatically manages deployments and runs tests before merges into master and pushes to branches.

The Jenkins *container* actually has full access to the Docker daemon on production. As a precaution against user error (but not necessarily making the system more secure), the jenkins user does not have direct root access to the server. Therefore, we have some pre-defined scripts `change-socket.docker` and `copy.docker` that run privledged to grant the dockerd unix socket and copy to /opt/telemetry-server on production. Note that since Jenkins has access to the Docker daemon, which runs as root, it should be treated as having root access.

Installing packages, running unit tests, and linting all occur in Docker containers as specified in the Jenkins pipeline. Actually copying the files occurs outside of the Docker containers
 
## Connecting to Jenkins

The Jenkins server in production can be accessed either at https://jenkins.solarracing.me or via an ssh tunnel.

To tunnel, you *must* connect to the Georgia Tech VPN (ssh traffic is only accessible within the Georgia Tech network, so eduroam/gtwifi won't work). Then, proxy-jenkins.sh will tunnel traffic on port 8080 to localhost:8088. *Note the final digit is different.*

Run

```
./proxy-jenkins.sh PRISM_ID
```
And then navigate to your web browser to localhost:8088 to access Jenkins jobs. We may in the future restrict Jenkins to Georgia Tech IPs via https://jenkins.solarracing.me
## Architecture

## Data Protocol

CAN data is first parsed by the telemetry subsystem on the car. The microcontroller connected to the Cellular LTE modem will convert each CAN frame into a basic 15-byte format, and forward it via TCP to port 6001 on the server.

| Bytes 0-3 | Bytes 4-5 | Bytes 6-7 | Bytes 8-15 |
|  ---      |  ---      |    ---    |   ---      |
|  'GTSR'   | CAN_ID    | Unused    | Payload    |

All data is transmitted in little-endian format. Can Frames that do not use the whole 8 byte payload must still transmit the fixed 8 byte size. Perhaps the unusued two bytes may be useful for a size specifier or a checksum in this instance.

## Handling connections

Each inbound TCP connection is assigned a `ConnectionHandler` struct in `listener.go`. We spin a different goroutine for every `ConnectionHandler` that each has a `packetParser` state machine that is responsible for processing the above data format.

## Parsing Data

Our CAN configurations are stored in `configs/can_config.json`. This JSON file stores the mappings between each metric and an offset/datatype for an appropriate CAN_ID. The parser state machines validates the data parsed and sees if the incoming CAN_ID contains a valid metric. If it does, extracts the approprate data, and forwards it to a global DatapointPublisher channel

A human readable table of all of our can configs can be found at `https://solarracing.me/data`

## Storage

All published datapoints are inserted into our influxdb database. We have a very simple schema which is only composed of the metric name, a timestamp, and a floating point value.

## Grafana

Grafana has built-in support for InfluxDB, and queries can be made via the Grafana dashboard. Note that a Datasource must be added that connects to http://influxdb:8086.

## Computations

In addition to metrics reported by the car, we provide functionality for small composable types to compute new metrics formed from the transformation of 1 or more other metrics. This is accomplished by subscribing to the datapoint publisher and fanning out a subset of relevant metric names to a channel to each computation client, which updates each computation client one data point at a time. Then, each computational client has the ability to write new metrics back into the global datapoint publisher channel, which can then be stored back into the storage layer.

## RF Listener

We interface with our RF subsystem by relaying our RF data to the tcp input of a server. The intention is that we can run all relevant parts of our server locallying on a laptop while trailering the car, and relay the RF data to localhost port 6001. We also provide the capability, if an internet connection is available, to relay the data to the server in production. This is primarily done via the Raspberry Pi in shop for debugging purposes.

The RF listener will also relay any chat messages to the car, to mirror functionality for LTE

Please see the README.md in rf-listener for more information on running the rf-listener. The rf-listener cannot be run in docker on OSX.

## Generator

For debugging purposes, we have a generator which will create Test Computations and Driver ACK/NACK messages and send them to either a local server or the production server. It will also print out any chat messages recieved from the server.
