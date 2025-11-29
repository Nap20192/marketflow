PORT=8080

down:
	docker-compose down -v

# Load exchange images and start services
up: load-exchanges
	docker-compose up -d

# Load exchange Docker images
load-exchanges:
	docker load -i ./generator/exchange1_amd64.tar
	docker load -i ./generator/exchange2_amd64.tar
	docker load -i ./generator/exchange3_amd64.tar

# Verify images are loaded
check-images:
	@echo "Checking if exchange images exist..."
	@docker images | grep -E "exchange[1-3]" || echo "Exchange images not found!"

run:
	go run cmd/main.go --port=$(PORT)

# Force rebuild and restart
rebuild: down load-exchanges
	docker-compose up --build -d

# Show logs for troubleshooting
logs:
	docker-compose logs -f

# Show status of all services
status:
	docker-compose ps
