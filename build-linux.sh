#!/bin/bash

BINARY=kbic-linux-amd64

# Build and test the Go project
echo "Building knowledge catalog Go application..."
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ${BINARY} cmd/kbase-catalog/main.go

if [ $? -eq 0 ]; then
    echo "Build successful!"
    
    # Show the binary exists
    if [ -f ${BINARY} ]; then
        echo "Executable created successfully: ${BINARY}"
    else
        echo "Error: Executable not found"
        exit 1
    fi
    
    # Show help
    echo ""
    echo "Usage examples:"
    echo "./${BINARY} process /path/to/root/directory"
    echo "./${BINARY} test /path/to/image.jpg"
else
    echo "Build failed!"
    exit 1
fi