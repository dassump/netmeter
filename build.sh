GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o netmeter-windows-amd64.exe .
GOOS=linux   GOARCH=amd64 go build -ldflags "-s -w" -o netmeter-linux-amd64       .