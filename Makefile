#!/usr/bin/make -f

# VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
# COMMIT := $(shell git log -1 --format='%H')

build:
	go build -o ./build/pond

install:
	go install