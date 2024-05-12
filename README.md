# Go-based presentation server for a react application

## TODO

### Backlog

- openapi
- github actions
- http session
- persistence
  - Real Mongodb collection
  - health integration
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
- Building (must)
  - helm
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

- service mesh

