#!/bin/bash

cd ..

CGO_ENABLE=0 GOOS=windows GOARCH=amd64 \
go build -o ./build/sw_nc_converter.exe main.go

