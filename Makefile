compose_up:
	@echo 'Run apps'
	docker-compose up --build -d

compose_down:
	@echo 'Stop running apps with docker-compose'
	docker-compose down