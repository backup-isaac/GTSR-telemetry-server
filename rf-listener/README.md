This package will listen to serial at the specified port for data, and immediately forward
the data to the TCP server at either solarracing.me, or localhost.

Note that the RF system does not do any pre-processing of the data; it sends it as-is to the
server.

NOTE: the below commands do not work for docker on OSX, as
the current implementation does not support hypervisor usb pass through
https://docs.docker.com/docker-for-mac/faqs/#can-i-pass-through-a-usb-device-to-a-container

## On Linux:
To run the rf-listener to send to the local server on docker:
```
docker build -t rf-listener rf-listener
docker run --rm=true --name=rf-listener -it --device=/dev/ttyUSB0 --network="telemetry-server_default" rf-listener go run listen.go /dev/ttyUSB0 server
```

For sending data to production run on docker:
```
docker build -t rf-listener rf-listener
docker run --rm=true --name=rf-listener -it --device=/dev/ttyUSB0  rf-listener go run listen.go /dev/ttyUSB0 remote
```
Alternatively, there is a convience Linux shell script:  `./listen-linux-docker.sh /dev/ttyUSB0 server`


## On OSX
Mac OSx Users must run the rf-listener locally/outside of docker since there is no way to pass through
the USB host

Send to local server
```
go run listen.go /dev/ttyUSB0 localhost
```
Send to production
```
go run listen.go /dev/ttyUSB0 remote
```
