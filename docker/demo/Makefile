.PHONY: all restart terminal logs;
all: restart
restart:
	docker compose pull demo
	docker compose stop demo
	find config -type f -name 'settings.yml' -delete
	docker compose up -d --force-recreate demo
install:
	docker compose pull
	docker compose stop
	docker compose up -d
terminal:
	docker compose exec demo bash
logs:
	docker compose logs --tail=50 -f
