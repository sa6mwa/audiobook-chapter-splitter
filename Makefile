BIN = bin
NAME = audiobook-chapter-splitter

.PHONY: all build install clean

all: build

build: bin
	CGO_ENABLED=0 go build -trimpath -o $(BIN)/$(NAME) .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -o $(BIN)/$(NAME)-darwin-arm64 .

install:
	install $(BIN)/$(NAME) /usr/local/bin/

clean:
	rm -rf $(BIN)

bin:
	mkdir -p $(BIN)
