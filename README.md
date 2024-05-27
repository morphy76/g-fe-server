# Go-based presentation server for a react application

## TODO

### Doing

- Broken MemoryRepository in the http contextualized stack
- Broken MongoRepository in the http contextualized stack

### Backlog

- Mongo client: pool options
- Logging: create a functional approach to logs, attributes and log propagation
- whats the vendor directory
- cloud friendly http session
  - shared store
  - memstore is not a cloud firendly way to handle the HTTP session, I would like to use mongo to reduce the number of integrated systems but, the mongostore recommended by Gorilla is not a top choice
- redis integration
-resusable artifact
- otel
- service mesh (istio)
- godoc
- openapi
- github actions
- authentication & authorization (must)
  - APIs to access HTTP session (token retrieval: header or session)
  - JWT authenticated APIs
- Accessibility (fun)
- multitenancy (must)
  - HTTP header tenant resolver ("done")
  - JWT tenant resolver
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

### Bugs

- App.tsx setSelected causes 3 queries: example, previous selected and new selected, expected is just new selected (must)

## What happened

Learning Go developing a presentation server: a presentation server is a Backend For Front-end component (<https://microservices.io/patterns/apigateway.html>) which exposes, in this case, a react application of a very simple entity CRUD.

In this scenario, the BFF also performs CRUD operations instead of acting as a gateway to downstream microservices.

### go application

So much freedom while structuring the project needs to be somehow tamed: <https://github.com/golang-standards/project-layout>.

The entry point of the presentation server is `cmd/main.go`.

It uses the following third party dependencies:

- Gorilla for the HTTP stack, it seems to be the best featured choice;
- Zerolog for logging, see later;
- The official MongoDB go driver to bind MongoDb 7 (go.mongodb.org/mongo-driver);
- ...no more, so far, but I guess I'll try to play also with Redis.

#### flags

Wow, Go has it natively!

I tried to figure out a sort of internal framework to avoid huge files, SRP infrigements, anyway evolving the file structure and the packages while learning Go.

- The `internal/cli` package provides builders per configurable integration, e.g. binding to the database and HTTP serving;
- It tests the environment which has the priority;
- Perform validation;
- Return a factory method to convert the flags into option types (package `internal/options`) to use to set up `context.Context` contextes to use downstream;
- It contains just the overall bindings, not domain specific items like the collection to use, which is in the internal implementation of the repository in the `example` package.

#### logging

Looking for G best practices, I gave a look at several online articles and I've found this one which helped me a lot: <https://betterstack.com/community/guides/logging/best-golang-logging-libraries/>. It compares several logging approaches for golang applications and Zerolog is my pick.

The presentation server starts at debug level, it's the minimal threshold for a cloud deployment to let application management. Collecting and filtering logs is the way to provide views on them, this topic will be enhanced by application tracing.

The trace level can be enabled through command flags and its expected audience is development support.

#### routing

TODO

#### mongo

TODO

TBV: is it the right way to propagate the repository?

- persistence
  - Real Mongodb collection *
  - health integration

### react application

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

### packaging

#### builder images

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

### Minikube

TODO

```shell
minikube start -p go --cpus=8 --memory=32g
eval $(minikube docker-env -p go)
make deploy
kubectl config use-context go
kubectl create ns fe
helm dependency build tools/helm/g-fe-server
helm upgrade --install -n fe fe-server tools/helm/g-fe-server
```

- Building (must)
  - helm

