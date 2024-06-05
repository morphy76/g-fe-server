-include deploy.env

# Set binary commands
GO := go
DOCKER := docker
NPM := npm
NODEMON := nodemon

# Set the flags
#GOFLAGS := -mod=vendor
LDFLAGS := -ldflags="-s -w"
GCFLAGS := -gcflags="-m -l"
TESTFLAGS := -v
# DOCKERBUILDFLAGS := --no-cache
NPMFLAGS := --no-audit --no-fund

# Define the target binary name
TARGET := g-fe-server
TARGET_FE := ./web/build
DOCKERFILE := ./tools/docker/Dockerfile
DEPLOY_TAG ?= g-fe-server:0.0.1

# Define the source files
SOURCES := ./cmd/main/serve.go

# Runtime args
MONGO_ARGS := -db=1 -db-mongo-url=mongodb://127.0.0.1:27017/go_db -db-mongo-user=go -db-mongo-password=go
SERVE_ARGS := -ctx=/fe -static=$(TARGET_FE) -host=localhost
OTEL_ARGS := -otel-enabled=false
OIDC_ARGS := -oidc-issuer=http://localhost:28080/realms/gfes -oidc-client-id=ps -oidc-client-secret=BA4eYsij3vDerLdQTRp6khSKWSDQWdLr -oidc-scopes=openid,profile,email,offline_access

# Define the build target
build:
	@$(GO) test $(TESTFLAGS) ./...
	@$(GO) build $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) -o $(TARGET) $(SOURCES)

watch:
	@$(NODEMON) --watch './**/*.go' --signal SIGTERM --exec $(GO) run $(GOFLAGS) $(LDFLAGS) $(SOURCES) $(SERVE_ARGS) -trace $(OTEL_ARGS) $(OIDC_ARGS)

watch-mongo:
	@$(NODEMON) --watch './**/*.go' --signal SIGTERM --exec $(GO) run $(GOFLAGS) $(LDFLAGS) $(SOURCES) $(SERVE_ARGS) -trace $(MONGO_ARGS) $(OTEL_ARGS) $(OIDC_ARGS)

#FE Build
build-fe:
	@$(NPM) $(NPMFLAGS) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui test
	@$(NPM) --prefix ./web/ui run build

watch-fe:
	@$(NPM) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui run watch

build-all: clean build build-fe

run:
	$(GO) run $(GOFLAGS) $(LDFLAGS) $(GCFLAGS) $(SOURCES) $(SERVE_ARGS) $(OTEL_ARGS) $(OIDC_ARGS)

run-mongo:
	$(GO) run $(GOFLAGS) $(LDFLAGS) $(SOURCES) $(SERVE_ARGS) $(MONGO_ARGS) $(OTEL_ARGS) $(OIDC_ARGS)

# Define the clean target
clean:
	-@rm -f $(TARGET)
	-@rm -rf $(TARGET_FE)

deploy: clean
	@$(DOCKER) run -d --network host --rm -v /var/run/docker.sock:/var/run/docker.sock --name socat alpine/socat tcp-listen:12345,fork,reuseaddr,ignoreeof unix-connect:/var/run/docker.sock
	-$(DOCKER) build --network host -t $(DEPLOY_TAG) -f $(DOCKERFILE) $(DOCKERBUILDFLAGS) .
	@$(DOCKER) stop socat
