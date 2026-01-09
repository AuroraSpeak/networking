
build-web:
	@echo "Building web..."
	cd web && npm run build

build-with-web: build-web
	@echo "Building with web..."
	go build -tags=debug -o bin/web ./cmd/web

build-server-headless:
	@echo "Building server headless..."
	go build -tags=debug -o bin/web ./cmd/web

run: build-server-headless
	@echo "Running headless..."
	./bin/web -tags=debug

dev-web:
	@echo "Running web in development mode..."
	cd web && npm run dev

run-with-web: build-with-web
	@echo "Running with web..."
	./bin/web -tags=debug

clean:
	rm -rf ./bin

make-server-headless:
	@echo "Building server headless..."
	go build -tags=debug -o bin/server ./cmd/server

run-server-headless: make-server-headless
	@echo "Running server headless..."
	./bin/server -tags=debug

make-client:
	@echo "Building client..."
	go build -tags=debug -o bin/client ./cmd/client

run-client: make-client
	@echo "Running client..."
	./bin/client -tags=debug

gen-docs:
	@echo "Generating docs..."
	go install golang.org/x/tools/cmd/godoc@latest

run-docs: gen-docs
	@echo "Running docs..."
	godoc -http=:6060

run-server-web-noui:
	@echo "Running server web without UI..."
	go build -tags=debug -o bin/web ./cmd/web
	./bin/web

.PHONY: build-web build-with-web dev-web run-with-web clean make-server-headless run-server-headless gen-docs run-docs run-client make-client