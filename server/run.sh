#!/bin/bash
export PATH=$PATH:/usr/local/go/bin
go mod tidy
go run ./cmd