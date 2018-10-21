This package will listen to serial at the specified port for data, and immediately forward
the data to the TCP server at either solarracing.me, or localhost.

Note that the RF system does not do any pre-processing of the data; it sends it as-is to the
server.

To run the rf-listener to localhost run:
$ docker exec -it server go run rf-listener/listen.go /dev/ttyUSB0

For sending data to the remote server run:
$ docker exec -it server go run rf-listener/listen.go /dev/ttyUSB0 remote
