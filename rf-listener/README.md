This package will listen to serial at the specified port for data, and immediately forward
the data to the TCP server at either solarracing.me, or localhost.

Note that the RF system does not do any pre-processing of the data; it sends it as-is to the
server.

To run the rf-listener to send to server run:
$ docker build -t rf-listener rf-listener
$ docker run --rm=true --name=rf-listener -it --device=/dev/ttyUSB0 --network="telemetry-server_default" rf-listener go run listen.go /dev/ttyUSB0 server

For sending data to the remote server run:
$ docker build -t rf-listener rf-listener
$ docker run --rm=true --name=rf-listener -it --device=/dev/ttyUSB0  rf-listener go run listen.go /dev/ttyUSB0 remote


