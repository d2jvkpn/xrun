BuildTime = $(shell date --iso=seconds)

build:	
	go build -ldflags="-X main._BuildTime=$(BuildTime)" -o xrun main.go
