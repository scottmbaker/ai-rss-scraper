PUBLISH_REPOSITORY             ?= smbaker
PUBLISH_REGISTRY               ?= docker.io
CHART_NAME                     ?= ai-rss-scraper
DOCKER_VERSION                 ?= 0.0.1
DOCKER_TAG                     ?= $(PUBLISH_REGISTRY)/$(PUBLISH_REPOSITORY)/$(CHART_NAME):$(DOCKER_VERSION)

.PHONY: build run fetch score test clean help

APP_NAME := ai-rss-scraper
CMD_PATH := ./cmd/$(APP_NAME)

build: ## Build the application
	go build -o bin/$(APP_NAME) $(CMD_PATH)

release: ## Build release binaries for Linux, Windows, and Mac
	GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME)-linux-amd64 $(CMD_PATH)
	GOOS=windows GOARCH=amd64 go build -o bin/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)
	GOOS=darwin GOARCH=amd64 go build -o bin/$(APP_NAME)-darwin-amd64 $(CMD_PATH)
	GOOS=darwin GOARCH=arm64 go build -o bin/$(APP_NAME)-darwin-arm64 $(CMD_PATH)

run: ## Run tool to fetch new articles and score them immediately
	go run $(CMD_PATH) run

fetch: ## Run tool to fetch new articles and save to DB
	go run $(CMD_PATH) fetch

score: ## Run tool to score unscored articles in DB
	go run $(CMD_PATH) score

list: ## List recent articles and their scores
	go run $(CMD_PATH) list

report: ## Generate report of articles
	go run $(CMD_PATH) report --out report.html

serve: ## Run the web server to list and manage articles
	go run $(CMD_PATH) serve

dump: ## Dump DB
	go run $(CMD_PATH) dump

test: ## Run tests
	go test ./...

lint: ## Lint the source code
	golangci-lint run

docker-build: ## Build docker image
	docker build -t $(PUBLISH_REPOSITORY)/$(CHART_NAME) .

docker-push: ## push helm chart to dockerhub
	docker tag $(PUBLISH_REPOSITORY)/$(CHART_NAME) $(DOCKER_TAG)
	docker push $(DOCKER_TAG)

clean: ## Clean build artifacts
	rm -rf bin/

help: ## Display this help screen
	@echo $(APP_NAME) make targets
	@echo "Target               Makefile:Line    Description"
	@echo "-------------------- ---------------- -----------------------------------------"
	@grep -H -n '^[[:alnum:]_-]*:.* ##' $(MAKEFILE_LIST) \
    | sort -t ":" -k 3 \
    | awk 'BEGIN  {FS=":"}; {sub(".* ## ", "", $$4)}; {printf "%-20s %-16s %s\n", $$3, $$1 ":" $$2, $$4};'
