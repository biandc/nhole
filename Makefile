export PATH := $(GOPATH)/bin:$(PATH)
export GO111MODULE=on
LDFLAGS := -s -w

all: clean fmt vet build

build: nhole-server nhole-client

fmt:
	go fmt ./...

vet:
	go vet ./...

nhole-server:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/nhole-server ./cmd/server/

nhole-client:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/nhole-client ./cmd/client/

clean:
	rm -f ./bin/nhole-server
	rm -f ./bin/nhole-client
	rm -f ./bin/nhole-server.exe
	rm -f ./bin/nhole-client.exe
