# Set binary commands
GO := go
DOCKER := docker
NPM := npm

# Set the flags
#GOFLAGS := -mod=vendor
LDFLAGS := -ldflags="-s -w"
TESTFLAGS := -v
# DOCKERBUILDFLAGS := --no-cache

# Define the target binary name
TARGET := g-fe-server
DOCKERFILE := ./tools/docker/Dockerfile
DEPLOY_TAG := g-fe-server:0.0.1

# Define the source files
SOURCES := ./cmd/main.go

all: clean test build

test:
	@$(GO) test $(TESTFLAGS) ./...

# Define the build target
build:
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(TARGET) $(SOURCES)

#FE Build
build-fe:
	@$(NPM) --prefix ./web/ui i
	@$(NPM) --prefix ./web/ui run build

run:
	@$(GO) run $(GOFLAGS) $(LDFLAGS) $(SOURCES) /fe ./web/build

# Define the clean target
clean:
	@rm -f $(TARGET)

deploy:
	@$(DOCKER) run -d --network host --rm -v /var/run/docker.sock:/var/run/docker.sock --name socat alpine/socat tcp-listen:12345,fork,reuseaddr,ignoreeof unix-connect:/var/run/docker.sock
	-$(DOCKER) build --network host -t $(DEPLOY_TAG) -f $(DOCKERFILE) $(DOCKERBUILDFLAGS) .
	@$(DOCKER) stop socat
