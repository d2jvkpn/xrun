BuildTime = $(shell date --iso=seconds)

build:	
	go build -ldflags="-X main.BuildTime=$(BuildTime)" -o xrun main.go
