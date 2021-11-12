#!/bin/bash
rm -rf docs
swag init -g communication/http/http.go
go build

if [[ $1 = "upload" ]]; then
	scp lilpop-server ubuntu@ec2:/opt/lilpop/
fi
