.DEFAULT_GOAL := help

SHELL := /usr/bin/env bash

ifeq ($(shell uname -s), Linux)
	OPEN := xdg-open
else
	OPEN := open
endif

project := tickers

.PHONY: setup # Setup a working environment
setup:
	@go mod download && go mod verify

.PHONY: lint # Run linter
lint:
	@staticcheck ./app/

.PHONY: prepare-test-containers
prepare-test-containers:
	@echo Starting db container...
	@docker run -d \
		--rm \
		--pull missing \
		--name ${project}_test_db \
		--tmpfs /var/lib/postgresql/data \
		--env-file ./.env.example \
		-p 5433:5432 \
		postgres:14-alpine

stop-prepared-test-containers := echo; \
	echo Stopping db container...; \
	docker stop ${project}_test_db

.PHONY: test # Run tests
test: prepare-test-containers
	@sleep 1
	@trap '${stop-prepared-test-containers}' EXIT && \
		echo Starting tests... && \
		dotenv -e .env.example -- env POSTGRES_PORT=5433 go test ./app/

.PHONY: coverage # Run tests with coverage report
coverage: prepare-test-containers
	@sleep 1
	@trap '${stop-prepared-test-containers}' EXIT && \
		echo Starting tests... && \
		dotenv -e .env.example -- env POSTGRES_PORT=5433 \
			go test -coverprofile=coverage.out ./app/ && \
		go tool cover -html=coverage.out -o coverage.html && \
		${OPEN} coverage.html

.PHONY: prepare-temp-containers
prepare-temp-containers:
	@echo Starting db container...
	@docker run -d \
		--rm \
		--pull missing \
		--name ${project}_temp_db \
		--tmpfs /var/lib/postgresql/data \
		--env-file ./.env.example \
		-p 5432:5432 \
		postgres:14-alpine

stop-prepared-temp-containers := echo; \
	echo Stopping db container...; \
	docker stop ${project}_temp_db

.PHONY: start # Start application
start: prepare-temp-containers
	@sleep 1
	@trap '${stop-prepared-temp-containers}' EXIT && \
		echo Starting application... && \
		dotenv -e .env.example -- go run ./app/

.PHONY: db
db: prepare-temp-containers
	@trap '${stop-prepared-temp-containers}' EXIT && \
		echo Press CTRL+C to stop && \
		sleep 1d

.PHONY: app
app:
	@echo Starting application... && \
		dotenv -e .env.example -- go run ./app/

.PHONY: up # Start Compose services
up:
	docker-compose -p ${project} -f docker-compose.yml pull db
	docker-compose -p ${project} -f docker-compose.yml build --pull
	docker-compose -p ${project} -f docker-compose.yml up

.PHONY: down # Stop Compose services
down:
	docker-compose -p ${project} -f docker-compose.yml down

.PHONY: help # Print list of targets with descriptions
help:
	@echo; \
		for mk in $(MAKEFILE_LIST); do \
			echo \# $$mk; \
			grep '^.PHONY: .* #' $$mk \
			| sed 's/\.PHONY: \(.*\) # \(.*\)/\1	\2/' \
			| expand -t20; \
			echo; \
		done
