.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2}'
	@echo ""

build-web: ## Build the web frontend
	@echo "Building web..."
	cd web && npm run build

build-with-web: build-web ## Build backend with web frontend
	@echo "Building with web..."
	go build -tags=debug -o bin/web ./cmd/web

build-server-headless: ## Build server headless
	@echo "Building server headless..."
	go build -tags=debug -o bin/web ./cmd/web

run: build-server-headless ## Run headless server
	@echo "Running headless..."
	./bin/web -tags=debug

dev-web: ## Run web frontend in development mode
	@echo "Running web in development mode..."
	cd web && npm run dev

run-with-web: build-with-web ## Run backend with web frontend
	@echo "Running with web..."
	./bin/web -tags=debug

clean: ## Clean build artifacts
	rm -rf ./bin
	rm networking_code_*.zip

make-server-headless: ## Build server headless binary
	@echo "Building server headless..."
	go build -tags=debug -o bin/server ./cmd/server

run-server-headless: make-server-headless ## Run server headless
	@echo "Running server headless..."
	./bin/server -tags=debug

make-client: ## Build client binary
	@echo "Building client..."
	go build -tags=debug -o bin/client ./cmd/client

run-client: make-client ## Run client
	@echo "Running client..."
	./bin/client -tags=debug

gen-docs: ## Generate documentation tools
	@echo "Generating docs..."
	go install golang.org/x/tools/cmd/godoc@latest

run-docs: gen-docs ## Run documentation server
	@echo "Running docs..."
	godoc -http=:6060

run-server-web-noui: ## Run server web without UI
	@echo "Running server web without UI..."
	go build -tags=debug -o bin/web ./cmd/web
	./bin/web

dev-backend: ## Run backend in development mode with air
	air

.PHONY: help build-web build-with-web dev-web run-with-web clean make-server-headless run-server-headless gen-docs run-docs run-client make-client run-server-web-noui dev-backend