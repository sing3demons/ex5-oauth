.PHONY: run build test clean install

run:
	go run main.go

build:
	go build -o oauth2-server main.go

test:
	go test -v ./...

clean:
	rm -f oauth2-server
	go clean

install:
	go mod download

docker-mongo:
	docker run -d -p 27017:27017 --name mongodb mongo:latest

docker-mongo-stop:
	docker stop mongodb && docker rm mongodb
