.PHONY: clean client server

all: server client

clean:
	rm -rf ./dist
	rm -f ./vue-ssr-v8go
	go clean

client:
	cd frontend && npm run build

server:
	go build -o vue-ssr-v8go .
