.PHONY: help up down restart logs ps clean

help:
	@echo "Comandos:"
	@echo "  make up       - Sobe containers"
	@echo "  make down     - Derruba containers"
	@echo "  make logs     - Logs de todos"
	@echo "  make restart  - Reinicia"
	@echo "  make ps       - Status"

up:
	docker-compose up -d

down:
	docker-compose down

restart:
	docker-compose restart

logs:
	docker-compose logs -f

ps:
	docker-compose ps

clean:
	docker-compose down -v --remove-orphans

.DEFAULT_GOAL := help