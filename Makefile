# Set binary commands
GO := go
DOCKER := docker
NPM := npm
NODEMON := nodemon

# Set the flags
#GOFLAGS := -mod=vendor
LDFLAGS := -ldflags="-s -w"
TESTFLAGS := -v
# DOCKERBUILDFLAGS := --no-cache

# Define the target binary name
TARGET := g-fe-server
TARGET_FE := ./web/build
DOCKERFILE := ./tools/docker/Dockerfile
DEPLOY_TAG := g-fe-server:0.0.1

# Define the source files
SOURCES := ./cmd/main.go

all: clean test build

test:
	@$(GO) test $(TESTFLAGS) ./...
	@$(NPM) --prefix ./web/ui test

# Define the build target
build:
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(TARGET) $(SOURCES)

watch:
	@$(NODEMON) --watch './**/*.go' --signal SIGTERM --exec $(GO) run $(GOFLAGS) $(LDFLAGS) $(SOURCES) /fe $(TARGET_FE)

#FE Build
build-fe:
	@$(NPM) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui run build

watch-fe:
	@$(NPM) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui run watch

run: clean build-fe
	@$(GO) run $(GOFLAGS) $(LDFLAGS) $(SOURCES) /fe $(TARGET_FE)

# Define the clean target
clean:
	-@rm -f $(TARGET)
	-@rm -rf $(TARGET_FE)

deploy: clean
	@$(DOCKER) run -d --network host --rm -v /var/run/docker.sock:/var/run/docker.sock --name socat alpine/socat tcp-listen:12345,fork,reuseaddr,ignoreeof unix-connect:/var/run/docker.sock
	-$(DOCKER) build --network host -t $(DEPLOY_TAG) -f $(DOCKERFILE) $(DOCKERBUILDFLAGS) .
	@$(DOCKER) stop socat
