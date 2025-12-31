
build-web:
	cd web && npm run build

build-with-web: build-web
	go build -o bin/web ./cmd/web

build-server-headless:
	go build -o bin/web ./cmd/web

run: build-server-headless
	./bin/web

dev-web:
	cd web && npm run dev

run-with-web: build-with-web
	./bin/web

clean:
	rm -rf ./bin

.PHONY: build-web build-with-web dev-web run-with-web clean