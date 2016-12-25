# SHELL := /bin/bash
APP_NAME := $(shell basename "$(CURDIR)")
BRANCH_NAME ?= $(shell git rev-parse --abbrev-ref HEAD)
NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

default: build

build: 
	@echo "$(OK_COLOR)+++++ Building binary $(APP_NAME) +++++$(NO_COLOR)"
	GOOS=linux go build -o ${APP_NAME} cmd/main.go 

image: build
	@echo "$(OK_COLOR)+++++ Building docker image $(APP_NAME) +++++$(NO_COLOR)"
	docker build -t $(APP_NAME) .

demo: image
	@echo "$(OK_COLOR)+++++ Running very simple tests in docker container +++++$(NO_COLOR)"
	docker run -it --rm $(APP_NAME)
