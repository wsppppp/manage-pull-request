# Makefile for the PR Reviewer Manager project

# .PHONY гарантирует, что make выполнит команду, даже если файл с таким именем существует.
.PHONY: help up down down-v build logs

# Определяем цель по умолчанию, которая будет выполняться, если запустить `make` без аргументов.
.DEFAULT_GOAL := help

help: ## Показать это справочное сообщение
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Собрать образы и запустить сервисы в фоновом режиме
	docker-compose up --build -d

down: ## Остановить сервисы
	docker-compose down

down-v: ## Остановить сервисы и удалить том с данными БД
	docker-compose down -v

build: ## Принудительно пересобрать образ приложения
	docker-compose build

logs: ## Показать логи сервисов и следить за ними
	docker-compose logs -f
