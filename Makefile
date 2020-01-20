hotels_server:
	go1.13 build -o bin/hotels-server cmd/hotels/main.go

sagas:
	go1.13 build -o bin/sagas cmd/sagas/main.go