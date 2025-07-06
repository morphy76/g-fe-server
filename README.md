# Go-based presentation server for a react application

## Known TODO

In summary:

- [x] With HTTP session management
- [x] With Health endpoint
  - Testing HTTP serving
  - Testing Mongo connection
- [ ] RBAC and resource server
- [ ] CORS options
  - Admin APIs with RBAC
- [ ] CSP options
  - Admin APIs with RBAC
- [x] With OIDC integration
- [x] With Mongo integration
- [ ] With Unleash integration
  - Feature gateway for FE flags
- [ ] Review monitoring (http clients)
  - MongoDB monitoring (Fix mongo monitoring (see `internal/db/monitor.go`))
  - OIDC monitoring
  - Unleash monitoring
  - add business events and errors to spans
- (With AIW logging)
- (With AIW tracing)
- (With AIW audit)
- With OpenAPI endpoint
- [x] No business logic included

Openpoints:

- Observability fallback to Opentracing/Jaeger using build flags
- Middleare and processing for tenant attribution
- Framework/standard to build the business request model: how logger, feserver, clients reaches the business module (*)
- A React 19 case study with MFEs
- TBV If Unleash is not enabled, we assume the feature is enabled by default
- KC onLogout:
  - 2025-07-06 20:35:17,521 WARN  [org.keycloak.events] (executor-thread-87) type="LOGOUT_ERROR", realmId="c304140e-bede-4b44-b5be-8ce28e637bea", realmName="gfes", clientId="ps", userId="null", ipAddress="0:0:0:0:0:0:0:1", error="invalid_redirect_uri", redirect_uri="/fe/ui"

## What happened

A presentation server is a Backend For Front-end component (<https://microservices.io/patterns/apigateway.html>) for react UIs.

### Go application

So much freedom while structuring the project needs to be somehow tamed: <https://github.com/golang-standards/project-layout>.

The entry point of the presentation server is `cmd/serve.go`.

It uses the following third party dependencies:

- Gorilla for the HTTP stack, it seems to be the best featured choice;
- Zerolog for logging, see later;
- The official MongoDB go driver to bind MongoDb 7 (go.mongodb.org/mongo-driver/v2);
- The official OpenTelemetry SDK for observability (go.opentelemetry.io/otel);
- Zitadel OIDC to bind the IAM (github.com/zitadel/oidc/v3);
- The official Unleash client for feature flags (github.com/Unleash/unleash-client-go);

#### flags

Wow, Go has it natively!

I tried to figure out a sort of internal framework to avoid huge files, SRP infrigements, anyway evolving the file structure and the packages while learning Go.

- The `cmd/cli` package provides builders per configurable integration, e.g. binding to the database and HTTP serving;
- It tests the environment which has the priority;
- Performs validation;
- Returns a factory method to convert the flags into option types (package `cmd/options`) to use to set up the FEServer;
- It contains just the overall bindings, not domain specific items.

#### Observability

The simplified-fe-server is integrated with the OpenTelemetry official SDK (`go.opentelemetry.io/otel`).

Additional integrations to observe third party dependencies enrich the spans providing instrumented HTTP clients to infrastructural dependencies.

The server also provides an enriched HTTP Client to propagate to the downstream (backend) services the trace context.

It exposrts the span, with a close to the default configuration, to Zipking which has been integrated as an Helm dependency.

#### Logging

Looking for Go best practices, I gave a look at several online articles and I've found this one which helped me a lot: <https://betterstack.com/community/guides/logging/best-golang-logging-libraries/>. It compares several logging approaches for golang applications and Zerolog is my pick.

The presentation server starts at debug level, it's the minimal threshold for a cloud deployment to let application management. Collecting and filtering logs is the way to provide views on them, this topic will be enhanced by application tracing.

The trace level can be enabled through command flags and its expected audience is development support.

Logging is enriched by contextual information:

- an _ownership_ dictionary to trace the attribution of the operation, in particular logical user organizations like tenants, subscriptions and stuff like that;
- a _correlation_ dictionary to trace the operation correlation, using the OTEL SDK, across decoupled or hierarchical operations: the log is enriched with the span_id and the trace_id;
- a _timing_ dictionary to trace the operation duration, in particular the time spent in the HTTP stack and in the business logic;
- a _routing_ dictionary to trace the GO runtime information;
- a _routine_ dictionary to trace the goroutine information, in particular the goroutine id.

#### Routing

Routing is hierarchical, `cmd/serve.go` prepares the server context and moves on to `internal/http/handlers/http_main_handler.go` to build the hierarchy.

Core handlers are in the `internal/http/handlers` package:

- `auth` for authentication routes: login, login callback, RP initiated logout logout and back-channel logout;
- `health` for health probes;
- `openapi` to serve the OpenAPI specification which is statically documented in the `api` package;
- `static` to serve the static content of the front end application.

#### Infrastructural dependencies, functional modules and dependency injection

The FEServer struct is in the `internal/server` package and it provides:

- HTTP serving options;
- Bindings to MongoDB for:
  - session store,
  - mongo client,
  - mongo database;
- the OIDC RelyingParty and ResourceServer;
- a facade to backend (AIW) services;
- a master switch to bind the feature server.

A functional module will be added under the `internal/business` package to provide business logic, It can inject dependencies extracting the FEServer from the request context, hence it should be structured in the following way:

```shell
internal/
└── business/
    ├── module1/
    │   ├── mod1_service.go
    │   ├── mod1_model.go
    │   ├── mod1_handlers.go
    │   └── mod1_*.go
    └── module2/
        └── ...
```

Hence this module can provide request or application scoped entities building from the related context which provides the FEServer.

As a best practice, it should be created as a request scoped module, see the example module (*).

A special module is the `internal/aiw` package prividing facilities and models to bind AIW platform services.

### React application

TODO + MFE

#### Webpack

TODO

- update FE resource on save (webpack watch)
- packaging within limits

#### FE stuff

TODO

- SASS
- favicon
- module path alias
- Internationalization react-intl

#### CRUD FE

TODO

- react hook form
- Add error handling
- react-query

#### FE Quality

TODO

- tslint

- continuous testing
- coverage

- reporting

### Packaging

#### Builder images

TODO

## Dependency upgrades

- Go mod TODO
- Frontend: `npm --prefix=./web/ui run dup`

## Run

### Development

- make watch-server
- make watch-fe

### Docker

- make deploy

### Helm and minikube

TODO

```shell
minikube start -p go --cpus=8 --memory=32g --kubernetes-version=v1.27.3
eval $(minikube docker-env -p go)
make deploy
kubectl config use-context go
kubectl create ns fe
helm dependency update tools/helm/g-fe-server
helm dependency build tools/helm/g-fe-server
helm upgrade --install -n fe fe-server tools/helm/g-fe-server

helm uninstall -n fe fe-server

kubectl -n fe port-forward services/fe-server-g-be-example 8081:8080
kubectl -n fe port-forward services/fe-server-g-fe-server 8080:8080
kubectl -n fe port-forward services/fe-server-zipkin 9411:9411
kubectl -n fe port-forward services/fe-server-prometheus-server 18080:80
```
