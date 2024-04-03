export PATH := $(GOPATH)/bin:$(PATH)

build:
	go mod tidy
	GOOS=linux GOARCH=amd64  go build -o ./bin/gl-socks5-demo ./


arm7:
	go mod tidy
	GOOS=linux GOARCH=arm  go build -o ./bin/gl-socks5-demo ./


arm64:
	go mod tidy
	GOOS=linux GOARCH=arm64  go build -o ./bin/gl-socks5-demo ./