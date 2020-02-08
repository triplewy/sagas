greeter_server:
	go build -o bin/greeter-server cmd/greeter/main.go
	
hotels_server:
	go build -o bin/hotels-server cmd/hotels/main.go

sagas:
	go build -o bin/sagas cmd/sagas/main.go