hotels_server:
	go build -o bin/hotels-server cmd/hotels/main.go

sagas:
	go build -o bin/sagas cmd/sagas/main.go