This package will target port 6001 and send formatted packets to mimic the car. All of the packets will have CAN id 0xFFFF, the Test metric. To test that the full data pipeline is working, just run data_generator.go with the server already running and see if data is being logged into the Test metric in Influx.

This program will also listen for uploaded track data from the map API. It will print the values that it received.

To run the generator on Mac or Linux, run generator.sh.

To run on Windows, you will need Go installed on your system. Run "go run data_generator.go"

This will target a locally running instance of the server.

To target the remote server, add the argument "remote" (e.g. "./generator.sh remote").
