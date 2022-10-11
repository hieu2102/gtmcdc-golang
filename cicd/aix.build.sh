#!/bin/bash
export GOOS=aix
export GOARCH=ppc64
go build ./cmd/cdcfilter
# chmod 777 cdcfilter
# scp cdcfilter 10.10.11.162:/usr/bin/