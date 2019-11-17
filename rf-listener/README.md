This package will listen to serial at the specified port for data, and immediately forward
the data to the TCP server at either solarracing.me, or localhost.

# The RF Antennae and Modem
There are two RF antennae that are connected to the computer serving as the RF listener via the RF modem. One is a dipole antenna, connected directly to the right port of the modem. The other is a cloverleaf antenna, contained within a black box and connected to the left port of the modem. The modem itself is connected to the computer via a USB to UART converter. See the image below:
![Image of RF antennae and modem](https://i.imgur.com/CTeo8RW.jpg)

# The RF Listener
This is a script that listens for communications via a serial port, then ports them as-is to either a locally hosted server or the remote server at https://grafana.solarracing.me. No pre-processing of the data is done by the RF system before sending the data to the server.
*NOTE: Running directly via docker does not work on OSX as the current implementation does not support hypervisor usb pass through: https://docs.docker.com/docker-for-mac/faqs/#can-i-pass-through-a-usb-device-to-a-container.*

Follow the directions below to start the RF listener script, based on the OS you are using. Note that `/dev/ttyUSB0` and `COM4` are placeholder names for the RF modem's serial port; you will have to confirm what your computer lists the modem as.
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
