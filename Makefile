ENVIRONMENTS=infra/dev.env
include ${ENVIRONMENTS}

.PHONY: up down config ps build

up:
	@docker compose \
		--env-file ${ENVIRONMENTS} \
		--project-directory ${PWD} \
		up -d $(filter-out $@,$(MAKECMDGOALS))

ps:
	@watch docker compose \
		--env-file ${ENVIRONMENTS} \
		--project-directory ${PWD} \
		ps -a
restart:
	@docker compose \
		--env-file ${ENVIRONMENTS} \
		--project-directory ${PWD} \
		restart	

down:
	@docker compose --env-file ${ENVIRONMENTS} \
		--project-directory ${PWD} \
		 down

config:
	@docker compose \
		--env-file ${ENVIRONMENTS} \
		--project-directory ${PWD} \
		config

build:
	@docker compose \
		--env-file ${ENVIRONMENTS} \
		--project-directory ${PWD} \
		--project-name marketplace \
		build $(SERVICE) --no-cache

psql:
	@psql -h localhost -U ${POSTGRES_USER} -d ${POSTGRES_DB} -p 5432
