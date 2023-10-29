#!/bin/bash

echo "Compiling.."
go build -o veil.exe main.go

if [ $? -ne 0 ]; then
    echo "Failed to compile"
    exit 1
fi

echo "Compiled successfully"
exit 0