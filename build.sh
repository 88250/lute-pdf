#!/bin/bash

go version
export GO111MODULE=on
export GOPROXY=https://goproxy.cn

go build -v
