# Build the Go binary locally
.PHONY: build
build:
	go build -o ledtx .

# Build the Docker container
.PHONY: build-container
build-container:
	./build/build.sh

# Run the Docker container
.PHONY: run
run:
	./build/run.sh

# Install the Docker container as a persistent service
.PHONY: install
install:
	./build/install.sh

# Clean build artifacts
.PHONY: clean
clean:
	rm -f ledtx
