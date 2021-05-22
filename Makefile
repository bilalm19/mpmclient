GO_BUILD = go build -o
EXECUTABLE = mpmclient
CGO = 0

.PHONY: build
build:
	CGO_ENABLED=$(CGO) $(GO_BUILD) $(EXECUTABLE) main.go
