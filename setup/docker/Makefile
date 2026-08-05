MAKEFLAGS += --always-make

-include .env

# Ensure compatibility with "docker-compose" (old) and "docker compose" (new).
HAS_DOCKER_COMPOSE_WITH_DASH := $(shell which docker-compose)

ifdef HAS_DOCKER_COMPOSE_WITH_DASH
    DOCKER_COMPOSE=docker-compose
else
    DOCKER_COMPOSE=docker compose
endif

# Overview of the available targets, shown by "make help".
define HELP_TEXT
PhotoPrism Docker Makefile

Usage: make <target> [target ...]

Manage:
  start            Start PhotoPrism in the background
  stop             Stop PhotoPrism
  restart          Stop and start PhotoPrism
  pull             Download the latest images
  update           Download the latest images and restart PhotoPrism
  logs             Follow the log output, press Ctrl+C to stop
  terminal         Open a terminal in the photoprism container

Remove (these delete data, so make sure you have a backup):
  reset            Reset the index, clear the cache, and remove sidecar files
  down             Stop the services and remove their containers and volumes
  remove           Remove the stopped containers and their volumes
  uninstall        Stop the services, then remove the containers and volumes

Running "make" without a target downloads the latest images and restarts.
endef
export HELP_TEXT

all: pull restart
help:
	@printf '%s\n' "$$HELP_TEXT"
uninstall: down remove
restart: stop start
update: pull restart
pull:
	$(DOCKER_COMPOSE) pull --ignore-pull-failures
start:
	$(DOCKER_COMPOSE) up -d --remove-orphans
stop:
	$(DOCKER_COMPOSE) stop
reset:
	$(DOCKER_COMPOSE) exec photoprism photoprism reset
down:
	$(DOCKER_COMPOSE) down -v
remove:
	$(DOCKER_COMPOSE) rm -s -v
terminal:
	$(DOCKER_COMPOSE) exec photoprism bash
logs:
	$(DOCKER_COMPOSE) logs --tail=50 -f
