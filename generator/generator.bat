docker run --rm -t -i --name generator -v %CD%:/app --network="telemetry-server" golang:1.11.2 go run /app/data_generator.go %1 %2