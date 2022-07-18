.PHONY: postgres-up
postgres-up:
	docker-compose -f docker-compose.yml up --build -d postgres

.PHONY: postgres-down
postgres-down:
	docker-compose -f docker-compose.yml stop postgres

.PHONY: all-down
all-down:
	docker-compose -f docker-compose.yml stop
	docker system prune

.PHONY: app-down
app-down:
	docker-compose -f docker-compose.yml stop app

.PHONY: app-up
app-up:
	docker-compose -f docker-compose.yml up -d app

.PHONY: all
all:
	docker-compose -f docker-compose.yml up -d