.PHONY: clean web server

all: server web

clean:
	rm -rf ./dist
	rm -f ./vue-ssr-v8go
	go clean

web:
	cd frontend && npm run build

server:
	go build -o vue-ssr-v8go .
