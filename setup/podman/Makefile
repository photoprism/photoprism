MAKEFLAGS += --always-make

-include .env

# Overview of the available targets, shown by "make help".
define HELP_TEXT
PhotoPrism Podman Makefile

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
	podman-compose pull
start:
	podman-compose up -d --remove-orphans
stop:
	podman-compose stop
reset:
	podman-compose exec photoprism photoprism reset
down:
	podman-compose down -v
remove:
	podman-compose rm -s -v
terminal:
	podman-compose exec photoprism bash
logs:
	podman-compose logs --tail=50 -f
