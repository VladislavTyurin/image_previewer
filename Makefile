build:
	docker compose build
	
run:
	docker-compose up -d

lint:
	golangci-lint -v run ./...

test:
	go test -v -count=10 -race -timeout=1m ./...

integration_test:
	cd integration_tests && docker-compose up -d
	cd integration_tests && go test --tags=integration -v -count=10 -race -timeout=2m ./
	cd integration_tests && docker-compose down