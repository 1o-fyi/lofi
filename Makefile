CC := GO111MODULE=on CGO_ENABLED=0 go
CFLAGS := build -o
SHELL := /bin/bash

# These are the only values you need to configure
# after you've change them you can run make rename to propogate your changes
NAME := lofi
USER := johns
HOST := git.sr.ht
UPSTREAM := github.com

CAP-ALL := cfg/all-caps.cfg
CAP-BIND := cfg/bind-cap.cfg
URL := UNAVAILABLE

MODULE := $(HOST)/~$(USER)/$(NAME)
REMOTE := git@$(HOST):~$(USER)/$(NAME)

DAEMON_CONFIG := $(NAME).service
DAEMON_ENV := /etc/conf.d/$(NAME)
DAEMON_PATH := /var/local/$(NAME)
DAEMON_CONFIG_PATH := /etc/systemd/system/$(DAEMON_CONFIG)
VERSION := $(shell ./tag)

build :: copy-local

# This will rename everything && create a new go module 
rename ::
				$(shell rm -rf go.* && go mod init $(MODULE) && go mod tidy && cp default.service $(NAME).service && git remote remove origin && git remote add origin $(REMOTE))

remote ::
			   	@echo $(REMOTE)

module ::
				@echo $(MODULE)

init :: build
				$(CC) mod init $(MODULE)

mod-install :: 
				$(CC) install ./... 

tidy :: mod-install
				$(CC) mod tidy -compat=1.17
				
format :: tidy
				$(CC)fmt -w -s *.go cmd/*.go

test ::	 format
				$(CC) test -v ./...

compile :: test
				$(CC) $(CFLAGS) $(MODULE) && chmod 755 $(MODULE)

link-local :: compile
				$(shell ldd $(MODULE))

headers :: link-local
				$(shell readelf -h $(MODULE) > $(MODULE).headers)

copy-local :: headers
				cp $(MODULE) .

send ::
				cd .. && tar cf $(NAME).$(VERSION).tar.xz $(NAME)/ && wormhole send $(NAME).$(VERSION).tar.xz   

get-tag ::
				$(shell curl https://git.sr.ht/~johns/tag/blob/main/tag > tag && chmod 755 tag)

get-go ::
				$(shell curl https://git.sr.ht/~johns/install-go/blob/main/install-go > install-go && chmod 755 install-go)

get-license ::
				$(shell curl https://www.gnu.org/licenses/agpl-3.0.txt > LICENSE)

install-scripts :: get-tag get-go

install :: install-scripts get-license

