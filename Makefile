.PHONY: client coordinator hotels

client:
	go build -o bin/client cmd/client/main.go
	
coordinator:
	go build -o bin/coordinator cmd/coordinator/main.go
	
hotels:
	go build -o bin/hotels cmd/hotels/main.go

