This package will listen to serial at the specified port for data, and immediately forward
the data to the TCP server at either solarracing.me, or localhost.

Note that the RF system does not do any pre-processing of the data; it sends it as-is to the
server.

NOTE: Running directly via docker does not work on OSX as
the current implementation does not support hypervisor usb pass through
https://docs.docker.com/docker-for-mac/faqs/#can-i-pass-through-a-usb-device-to-a-container

## On Linux:
To run the rf-listener to send to the local server on docker:
```
./listen-linux-docker.sh /dev/ttyUSB0 server
```

For sending data to production run on docker:
```
./listen-linux-docker.sh /dev/ttyUSB0 remote
```

## On OSX
Mac OSX and Windows users must either run or build locally. To build locally via docker, run `build.sh`

Send to local server
```
./bin/listen.app /dev/ttyUSB0 localhost
```
Send to production
```
./bin/listen.app /dev/ttyUSB0 remote
```

## On Windows
Mac OSX and Windows users must either run or build locally. To build locally via docker, run `.\build.bat` in Powershell

Send to local server
```
.\bin\listen.exe COM4 localhost
```
Send to production
```
.\bin\listen.exe COM4 remote
```
