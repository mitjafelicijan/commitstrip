default:
	@echo "Default target. Check list of targets for help."

run:
	mkdir -p download
	go run .

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/cslocal -v .
