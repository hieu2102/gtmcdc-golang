#!/bin/bash
export GOOS=linux
export GOARCH=amd64
go build ./cmd/cdcfilter
chmod 777 cdcfilter
scp cdcfilter 10.10.11.162:/usr/bin/