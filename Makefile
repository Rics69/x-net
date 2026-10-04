include .env
export

export PROJECT_ROOT= $(shell pwd)

env-up:
	@docker compose up -d xchat-postgres xchat-rabbitmq

env-down:
	@docker compose down xchat-postgres xchat-rabbitmq

env-cleanup:
	@read -p "Очистить volume файлы бд? Опасность утери данных. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down xchat-postgres xchat-rabbitmq port-forwarder && \
		rm -rf ${PROJECT_ROOT}/out/pgdata ${PROJECT_ROOT}/out/rabbitmq && \
		echo "Файлы очищены"; \
	else \
		echo "Очистка отменена"; \
	fi

env-port-forward:
	@docker compose up -d port-forwarder

env-port-close:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm xchat-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq  "$(seq)"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует параметр action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm xchat-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@xchat-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

logs-cleanup:
	@read -p "Очистить log файлы? Опасность утери логов. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		rm -rf ${PROJECT_ROOT}/out/logs && \
		echo "Файлы логов очищены"; \
	else \
		echo "Очистка логов  отменена"; \
	fi

xchat-run:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=localhost && \
	export RABBITMQ_HOST=localhost && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/xchat/main.go

# второй инстанс на :5051 рядом с первым - для проверки доставки событий через RabbitMQ
xchat-run-2:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=localhost && \
	export RABBITMQ_HOST=localhost && \
	export HTTP_ADDR=:5051 && \
	go run ${PROJECT_ROOT}/cmd/xchat/main.go

xchat-deploy:
	@docker compose up -d --build xchat xchat-2

xchat-undeploy:
	@docker compose down xchat xchat-2

ps:
	@docker compose ps

swagger-gen:
	@docker compose run --rm swagger \
		init \
		-g cmd/xchat/main.go \
		-o docs \
		--parseInternal \
		--parseDependency
