# Go-based presentation server for a react application

## TODO

### Doing

- otel
- service mesh (istio)

### Backlog

- godoc
- openapi
- github actions
- http session
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

### go application

#### flags

TODO

#### logging

TODO

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

