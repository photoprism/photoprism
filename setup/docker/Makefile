MAKEFLAGS += --always-make

-include .env
export

# Ensure compatibility with "docker-compose" (old) and "docker compose" (new).
HAS_DOCKER_COMPOSE_WITH_DASH := $(shell which docker-compose)

ifdef HAS_DOCKER_COMPOSE_WITH_DASH
    DOCKER_COMPOSE=docker-compose
else
    DOCKER_COMPOSE=docker compose
endif

all: pull restart
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
