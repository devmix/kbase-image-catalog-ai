#!/bin/bash

BINARY=kbic-windows-amd64.exe

# Build and test the Go project
echo "Building knowledge catalog Go application..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -ldflags "-s -w" -o ${BINARY} cmd/kbase-catalog/main.go

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