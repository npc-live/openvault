BINARY     = openvault
BUILD_DIR  = .

.PHONY: build clean install tidy

build:
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BINARY) .

install:
	CGO_ENABLED=0 go install .

tidy:
	go mod tidy

clean:
	rm -f $(BUILD_DIR)/$(BINARY)
