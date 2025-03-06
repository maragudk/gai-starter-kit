TAILWINDCSS_OS_ARCH := macos-arm64
#TAILWINDCSS_OS_ARCH := linux-x64

.PHONY: benchmark
benchmark:
	go test -tags sqlite_fts5 -bench=. ./...

.PHONY: build-css
build-css: tailwindcss
	./tailwindcss -i tailwind.css -o public/styles/app.css --minify

.PHONY: build-docker
build-docker: build-css
	docker build --platform linux/amd64,linux/arm64 .

.PHONY: cover
cover:
	go tool cover -html=cover.out

.PHONY: lint
lint:
	golangci-lint run

models/mxbai-embed-large-v1-f16.llamafile:
	mkdir -p models
	curl -sLO https://assets.maragu.dev/llm/mxbai-embed-large-v1-f16.llamafile
	chmod a+x models/mxbai-embed-large-v1-f16.llamafile

.PHONY: start
start: build-css
	go run -tags sqlite_fts5 ./cmd/app

.PHONY: start-embeddings
start-embeddings: models/mxbai-embed-large-v1-f16.llamafile
	./models/mxbai-embed-large-v1-f16.llamafile --server --listen localhost:8082 --v2

tailwindcss:
	curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-$(TAILWINDCSS_OS_ARCH)
	mv tailwindcss-$(TAILWINDCSS_OS_ARCH) tailwindcss
	chmod a+x tailwindcss

.PHONY: test
test:
	go test -tags sqlite_fts5 -coverprofile=cover.out -shuffle on ./...

.PHONY: watch-css
watch-css: tailwindcss
	./tailwindcss -i tailwind.css -o public/styles/app.css --watch
