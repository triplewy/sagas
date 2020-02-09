.PHONY: hotels

greeter_server:
	go build -o bin/greeter-server cmd/greeter/main.go
	
hotels:
	go build -o bin/hotels cmd/hotels/main.go

coordinator:
	go build -o bin/coordinator cmd/coordinator/main.go