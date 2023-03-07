MAKEFLAGS += --always-make

-include .env
export

all: pull restart
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
