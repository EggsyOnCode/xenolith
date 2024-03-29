build: 
	go build -o ./bin/xenolith

run: build
	./bin/xenolith

test: 
	go test -v ./...
