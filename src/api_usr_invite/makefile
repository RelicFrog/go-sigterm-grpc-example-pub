#!make
# ---------------------------------------------------------------------------------------------------------------------
# MAKEFILE for GoLang gRPC Service Stack (project aribor)
# ---------------------------------------------------------------------------------------------------------------------
# @purpose: base commands for handling service provisioning in aribor
# @info: please configure/setup your makefile.json first!
# ---------------------------------------------------------------------------------------------------------------------
# @author: Patrick Paechnatz <post@dunkelfrosch.com>
# @version: 1.0.0 <alpha>
# @createdAt: 2020-11-12
# @updatedAt: 2020-11-16
# ---------------------------------------------------------------------------------------------------------------------

define GetFromConfigByCategory
$(shell jq -r -e '[ .$(1).$(2) ] | .[]' makefile.json)
endef

define GetProfileFromConfigByEnv
$(shell jq -r -e '[ .terraform.env.$(1).$(2) ] | .[]' makefile.json)
endef

.ONESHELL:
.SHELL := /usr/bin/bash
.PHONY: help

# -- makefile internal config stack
RAND_FILE=makefile.sec
TOKEN_FILE=makefile.token
SCREEN_WIDTH=$(shell tput cols)
BOLD=$(shell tput bold)
RED=$(shell tput setaf 1)
GREEN=$(shell tput setaf 2)
CYAN=$(shell tput setaf 4)
DARKCYAN=$(shell tput setaf 6)
RESET=$(shell tput sgr0)
GRAY=$(shell tput setaf 7)
# -- project variable baseline
PROJECT:=$(call GetFromConfigByCategory,project,name)
PROJECT_LONG:=$(call GetFromConfigByCategory,project,name_long)
PROJECT_DESC:=$(call GetFromConfigByCategory,project,description)
PROJECT_UNIT:=$(call GetFromConfigByCategory,project,unit)
PROJECT_VERSION:=$(call GetFromConfigByCategory,project,version)
PROJECT_NAMESPACE:=$(call GetFromConfigByCategory,project,namespace)
PROJECT_ROOT_PATH:= $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
PROJECT_UNSUPPORTED_OS:=$(call GetFromConfigByCategory,project,unsupported_os)
PROJECT_TIME_CURRENT:=$$(date +'%Y%m%d-%H%M%S')
PROJECT_BIN_REQ=$(call GetFromConfigByCategory,project,packages_req)
PROJECT_BIN_CHK:=$(foreach exec,$(PROJECT_BIN_REQ),$(if $(shell which $(exec)),some string,$(error "[FATAL] missing project required binary/lib [$(exec)] in shell PATH! Please install [$(exec)]")))
PROJECT_PATH:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
PROJECT_RANDOM_TOKEN:="-"
PROJECT_RANDOM_ID:="0"
PROJECT_BUILD_VERSION:=$(shell git describe --tags --always)
PROJECT_BUILD_HASH:=$(shell git rev-parse --short HEAD)
PROJECT_BUILD_NAME:=$(shell basename "$(PWD)")

# -- golang related variables (old)
# GOBASE:=$(shell pwd)
# GOPATH:=$(GOBASE)/vendor:$(GOBASE)
# GOBIN:=$(GOBASE)/bin
# GOBASE_GLOBAL:=$$HOME/go
# GOPATH_GLOBAL:=$(GOBASE_GLOBAL)/vendor:$(GOBASE_GLOBAL)
# GOBIN_GLOBAL:=$(GOBASE_GLOBAL)/bin

# -- golang related variables (new)
GO111MODULE=on
GOPATH=$$HOME/go
GOBIN=$(GOPATH)/bin

GOFILES:=$(wildcard *.go)
LDFLAGS=-ldflags "-X=main.Version=$(PROJECT_BUILD_VERSION) -X=main.Build=$(PROJECT_BUILD_HASH)"
STDERR:=/tmp/.$(PROJECT_BUILD_NAME)-stderr.txt
PID:=/tmp/.$(PROJECT_BUILD_NAME).pid
MAKEFLAGS += --silent

ifeq (, $(shell which jq))
$(error "no jq in $(PATH), consider doing install jq! <exit>")
endif

ifeq ($(OS),Windows_NT)
    OS_IDENT_RAW += WIN32
    ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
        OS_IDENT_RAW += AMD64
    else
        ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
            OS_IDENT += AMD64
        endif
        ifeq ($(PROCESSOR_ARCHITECTURE),x86)
            OS_IDENT_RAW += IA32
        endif
    endif
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        OS_IDENT_RAW += LINUX
    endif
    ifeq ($(UNAME_S),Darwin)
        OS_IDENT_RAW += OSX
    endif
    UNAME_P := $(shell uname -p)
    ifeq ($(UNAME_P),x86_64)
        OS_IDENT_RAW += AMD64
    endif
    ifneq ($(filter %86,$(UNAME_P)),)
        OS_IDENT_RAW += IA32
    endif
    ifneq ($(filter arm%,$(UNAME_P)),)
        OS_IDENT_RAW += ARM
    endif
endif

makefile-init:
	@{ \
		set -e ;\
		LOGO=`cat ../../.github/logo.asc` ;\
		OS_IDENT=`echo "$(OS_IDENT_RAW)" | tr " " -` ;\
		echo "\n$(BOLD)$(GRAY)$$LOGO$(RESET)\n" ;\
		printf '%.0s-' {1..$(SCREEN_WIDTH)};printf '\n' ;\
		echo "$(BOLD)$(GRAY)Project Service Makefile for [ $(RESET)$(DARKCYAN)$(PROJECT_NAMESPACE)/$(PROJECT_UNIT)/$(PROJECT) v$(PROJECT_VERSION)$(RESET)$(BOLD)$(GRAY) ] (c) 2019-2020 RelicFrog LLC$(RESET)" ;\
		printf '%.0s-' {1..$(SCREEN_WIDTH)};printf '\n' ;\
		echo "\n$(GRAY)---[ NFO ]: [project/description]    == $(RESET)$(DARKCYAN)$(PROJECT_DESC)$(RESET)" ;\
		if [[ "$(OS_IDENT)" =~ "$(PROJECT_UNSUPPORTED_OS)" ]]; then \
			echo "$(BOLD)$(RED)OS not supported by this makefile-processor ... sorry :-/ $(OS_IDENT)$(RESET)" ;\
			SYS_ERROR=1 ;\
		fi ;\
		if [ ! -z $${SYS_ERROR} ] && [ $${SYS_ERROR} -eq 1 ]; then \
			echo "$(BOLD)system error! <exit>$(RESET)" ;\
			exit 9 ;\
		fi ;\
		if [ ! -z $${CFG_ERROR} ] && [ $${CFG_ERROR} -eq 1 ]; then \
			echo "$(BOLD)configuration error! <exit>$(RESET)" ;\
			exit 1 ;\
		fi ;\
		if [[ ! -a $(RAND_FILE) ]] ; then \
			echo $$RANDOM > $(RAND_FILE) ;\
		fi ;\
		PROJECT_RANDOM_ID=`cat $(RAND_FILE)` ;\
		PROJECT_RANDOM_TOKEN=$$OS_IDENT-$(PROJECT_TIME_CURRENT)-$$PROJECT_RANDOM_ID ;\
		GIT_VERSION=`git describe --tags --always` ;\
		GIT_BUILD=`git rev-parse HEAD` ;\
		echo $$PROJECT_RANDOM_TOKEN > $(TOKEN_FILE) ;\
		echo "$(GRAY)---[ NFO ]: [project/build/token]    == $(RESET)$(DARKCYAN)$$PROJECT_RANDOM_TOKEN$(RESET)" ;\
		echo "$(GRAY)---[ NFO ]: [project/build/name]     == $(RESET)$(DARKCYAN)$(PROJECT_BUILD_NAME)$(RESET)" ;\
		echo "$(GRAY)---[ NFO ]: [project/git/hash]       == $(RESET)$(DARKCYAN)$$GIT_BUILD$(RESET)" ;\
		echo "$(GRAY)---[ NFO ]: [project/git/version]    == $(RESET)$(DARKCYAN)$$GIT_VERSION$(RESET)" ;\
		echo "$(RESET)" ;\
		printf '%.0s-' {1..$(SCREEN_WIDTH)};printf '\n' ;\
		echo "$(RESET)" ;\
	}

##-- [ Basic Commands ] --

## List available tasks on this project
help: makefile-init
	@awk '{ \
		if ($$0 ~ /^.PHONY: [a-zA-Z\-\0-9]+$$/) { \
			helpCommand = substr($$0, index($$0, ":") + 2); \
			if (helpMessage) { \
				printf "\033[36m%-20s\033[0m %s\n", \
					helpCommand, helpMessage; \
				helpMessage = ""; \
			} \
		} else if ($$0 ~ /^[a-zA-Z\-\0-9.]+:/) { \
			helpCommand = substr($$0, 0, index($$0, ":")); \
			if (helpMessage) { \
				printf "\033[36m%-20s\033[0m %s\n", \
					helpCommand, helpMessage; \
				helpMessage = ""; \
			} \
		} else if ($$0 ~ /^##/) { \
			if (helpMessage) { \
				helpMessage = helpMessage"\n                     "substr($$0, 3); \
			} else { \
				helpMessage = substr($$0, 3); \
			} \
		} else { \
			if (helpMessage) { \
				print "\n$(GRAY)"helpMessage"$(RESET)\n" ;\
			} \
			helpMessage = ""; \
		} \
	}' \
	$(MAKEFILE_LIST)

##-- [ GOLANG Commands ] --

## Install missing dependencies. Runs `go get` internally. e.g; make install get=github.com/foo/bar
install: go-get

## Run server in test mode
start: makefile-init compile
	@echo " *** starting gRPC server/service"
	@@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go run -mod=readonly server.go

## Compile the binary.
compile: makefile-init pbgen
	@-$(MAKE) -s go-compile

## Test the binary.
test: makefile-init
	@-$(MAKE) -s go-test

## Build docker image for service binary.
image: makefile-init
	@-$(MAKE) -s go-build-docker-img

## Run given command, wrapped with custom GOPATH. e.g; make exec run="go test ./..."
exec: makefile-init
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) $(run)

## Clean build files. Runs `go clean` internally.
clean: makefile-init
	@-rm bin/$(PROJECT_BUILD_NAME) 2> /dev/null
	@-$(MAKE) go-clean

##-- [ MONGODB Commands ] --

## Load primary service fixtures
dbfixtures: makefile-init
	@@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go run -mod=readonly server.go --load-fixtures

## Generate local mongo-db environments for dev/test
dbgen: makefile-init
	@$(shell [[ $(docker ps --filter "name=^/aribor-sys-mongodb-test$" --format '{{.Names}}') == aribor-sys-mongodb-test ]] || docker run -d -p 27018:27017 -e MONGODB_DATABASE=example_user_db --name rf-sys-mongodb-test bitnami/mongodb:4.4.1)

##-- [ PROTOBUF Commands ] --

## Generate primary service protobuf structure

pbgen: makefile-init
	@echo " *** Building protobuff classes"
	@PATH=$$PATH:$(GOPATH):$(GOBIN) protoc --go_out=plugins=grpc:proto -I ../../pb ../../pb/rf_example.proto

go-compile: go-get go-build

go-build:
	@echo " *** Building binary ..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go build $(LDFLAGS) -mod=readonly -o bin/$(PROJECT_BUILD_NAME) $(GOFILES)

go-generate:
	@echo " *** Generating dependency files ..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go generate $(generate)

go-get:
	@echo " *** Checking if there is any missing dependencies ..."
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go get $(get)

go-install:
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go install $(GOFILES)

go-test:
	@echo " *** Running all service related tests ..."
	./build/do_tests.sh

go-build-docker-img:
	@echo " *** Build docker image for $(PROJECT_LONG) ..."
	./build/do_img.sh

go-clean:
	@echo " *** Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean
