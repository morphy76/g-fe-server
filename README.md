# Go-based presentation server for a react application

## TODO

### Doing

- authentication & authorization (must)
  - APIs to access HTTP session (token retrieval: header or session)
  - JWT authenticated APIs

### Backlog

- multitenancy (must)
  - HTTP header tenant resolver ("done")
  - JWT tenant resolver
- helm review & service mesh (istio)
- a new service with Pact
  - e.g. (https://medium.com/@dees3g/pact-and-go-contract-testing-of-http-based-applications-e595e639334e)
- redis integration (as a client, as a mongo cache, as an http session store) + health
- kafka integration... mmm SSE/WS + frontend pseudo-chat (?) + health (sarama), anyway more protocols
- zookeeper playground?
- otel for system dependencies: mongo, kafka & redis
- prometheus for system dependencies: mongo, kafka & redis
- cloud friendly http session
  - shared store
  - memstore is not a cloud firendly way to handle the HTTP session, I would like to use mongo to reduce the number of integrated systems but, the mongostore recommended by Gorilla is not a top choice
- openapi
- godoc
- drill down about tests, e.g. fail when coverage is not achieved
- Logging: create a functional approach to logs, attributes and log propagation
- what's the vendor directory
- resusable artifact: pluggable domain resources, API & FE
- Accessibility (fun)
- FE crud
  - use query cache and optimistic updates (must)
- selected example in URL instead of react useState (fun)
  - uri driven component status
- error boundary (fun)
- storybook (fun)
- typedoc (fun)
- PWA (fun)
- improve responsiveness for mobiles? (fun)
- microfrontend serve layout (nice)
  - mfe out-of-the-box versioning
- React 19 (fun)
- more mongo client options

## What happened

Learning Go while developing a presentation server: a presentation server is a Backend For Front-end component (<https://microservices.io/patterns/apigateway.html>) which exposes, in this case, a react application of a very simple entity CRUD.

In this scenario, the BFF also performs CRUD operations instead of acting as a gateway to downstream microservices.

### Go application

So much freedom while structuring the project needs to be somehow tamed: <https://github.com/golang-standards/project-layout>.

The entry point of the presentation server is `cmd/main.go`.

It uses the following third party dependencies:

- Gorilla for the HTTP stack, it seems to be the best featured choice;
- Zerolog for logging, see later;
- The official MongoDB go driver to bind MongoDb 7 (go.mongodb.org/mongo-driver);
- The official OpenTelemetry SDK for observability (go.opentelemetry.io/otel);
- ...no more, so far, but I guess I'll try to play also with Redis.

#### flags

Wow, Go has it natively!

I tried to figure out a sort of internal framework to avoid huge files, SRP infrigements, anyway evolving the file structure and the packages while learning Go.

- The `internal/cli` package provides builders per configurable integration, e.g. binding to the database and HTTP serving;
- It tests the environment which has the priority;
- Perform validation;
- Return a factory method to convert the flags into option types (package `internal/options`) to use to set up `context.Context` contextes to use downstream;
- It contains just the overall bindings, not domain specific items like the collection to use, which is in the internal implementation of the repository in the `example` package.

#### Observability

The g-fe-server is integrated with the OpenTelemetry official SDK (`go.opentelemetry.io/otel`).

As a fake BFF (it's not acting as a gateway but having CRUD operations on `examples` directly on board), the spans are _local_.

Additional integrations to observe third party dependencies like MongoDB will enrich the spans.

It exposrts the span, with a close to the default configuration, to Zipking which has been integrated as an Helm dependency.

#### Logging

Looking for Go best practices, I gave a look at several online articles and I've found this one which helped me a lot: <https://betterstack.com/community/guides/logging/best-golang-logging-libraries/>. It compares several logging approaches for golang applications and Zerolog is my pick.

The presentation server starts at debug level, it's the minimal threshold for a cloud deployment to let application management. Collecting and filtering logs is the way to provide views on them, this topic will be enhanced by application tracing.

The trace level can be enabled through command flags and its expected audience is development support.

Logging is enriched by contextual information:

- an _ownership_ dictionary to trace the attribution of the operation, in particular logical user organizations like tenants, subscriptions and stuff like that;
- a _correlation_ dictionary to trace the operation correlation, using the OTEL SDK, across decoupled or hierarchical operations: the log is enriched with the span_id and the trace_id.

#### Routing

Routing is hierarchical, `cmd/main.go` prepares the server context and moves on to `internal/http/handlers/handler.go` to build the hierarchy.

First of all, the parent router which receives 3 middlewares:

- The one which sets the request context using values picked from the server context like, in particular to reuse pooled resources across concurrent requests:
  - The connected db client,
  - The session store;
- A tenant resolver, more on this later on...;
- A semi-pre-configured logger.

Then follows the context router, namely the router which responds to the application context path as configured by run arguments.

The next routers are task-focused:

- The static router to deliver static content, within an HTTP session, with some focus on routing single page applications;
- The API router to expose the BFF endpoints, but in this case they are the actual CRUD operations on the available resources;
- A non-functional router, in this case it provides the health probe for k8s environments.

Finally, waiting to learn how to plug stuff into a Go runtime, an hardcoded router to handle the _example_ resource.

Generally speaking, handle functions are provided by the router provided by each module, e.g. `internal/http/health/handler.go` has the health handle functions and `internal/example/http/handlers.go` has those related to the _example_ resource.

Routers, the API router in particular, are integrated with Opentracing with a Gorilla extension: `go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux`.

In the same way, such routers are configured to use Prometheus middlewares (`github.com/prometheus/client_golang`) to expose metrics about their usage.

#### MongoDB & domain repository

The database connection client and the domain repository are kept separated so that:

- The client can be connected at application level, in the application context;
- The repository can be in the request context.

Such entries of the application context are propagated to the request context by the parent router middleware, it is then up to the domain model use them to create domain artifacts.

This is shown in the _example_ HTTP handlers where, through the `ContextualizedApi` function, a _repository-enriched_ handler function is bound to the item router.

MongoDB is connected using the official library (`go.mongodb.org/mongo-driver`) and participate (synchronously so far) to the helth probe.

### React application

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

- make watch
- make watch-fe

### Docker

TODO

### Helm and minikube

TODO

```shell
minikube start -p go --cpus=8 --memory=32g
eval $(minikube docker-env -p go)
make deploy
kubectl config use-context go
kubectl create ns fe
helm dependency build tools/helm/g-fe-server
helm upgrade --install -n fe fe-server tools/helm/g-fe-server

helm uninstall -n fe fe-server

kubectl -n fe port-forward services/fe-server-g-fe-server 8080:8080
kubectl -n fe port-forward services/fe-server-zipkin 9411:9411
kubectl -n fe port-forward services/fe-server-prometheus-server 18080:80
```
