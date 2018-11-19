BIN=explainer

HEAD=$(shell git describe --tags 2> /dev/null  || git rev-parse --short HEAD)

all: build

build: darwin64 linux64 windows64

clean:
	-rm -f $(BIN)
	-rm -rf release

darwin64:
	env GOOS=darwin GOARCH=amd64 go clean -i
	env GOOS=darwin GOARCH=amd64 go build -o release/darwin64/$(BIN) .

linux64:
	env GOOS=linux GOARCH=amd64 go clean -i
	env GOOS=linux GOARCH=amd64 go build -o release/linux64/$(BIN) .

windows64:
	env GOOS=windows GOARCH=amd64 go clean -i
	env GOOS=windows GOARCH=amd64 go build -o release/windows64/$(BIN).exe .

.PHONY: release
release: clean build
	zip -9 release/$(BIN).darwin_amd64.$(HEAD).zip release/darwin64/$(BIN)
	zip -9 release/$(BIN).linux_amd64.$(HEAD).zip release/linux64/$(BIN)
	zip -9 release/$(BIN).windows_amd64.$(HEAD).zip release/windows64/$(BIN).exe
