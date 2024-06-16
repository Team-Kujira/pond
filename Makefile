#!/usr/bin/make -f

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

ldflags = -X pond/cmd.Version=$(VERSION) \
		  		-X pond/cmd.Commit=$(COMMIT)

BUILD_FLAGS := -ldflags '$(ldflags)'

build:
	go build $(BUILD_FLAGS) -o ./build/pond

install:
	go install $(BUILD_FLAGS)