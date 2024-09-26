
default:
  just --list

build-tw-css:
	npx tailwindcss -i input.css -o static/css/tw.css

build-templ:
  templ generate

build-webserver: build-tw-css build-templ (go-build "webserver")

build-github-workflows-import-schema:
  # NOTE: this requires cue v0.11.0-alpha.2 or later
  curl -o internal/ci/github/github.actions.workflow.schema.json https://raw.githubusercontent.com/SchemaStore/schemastore/f728a2d857a938979f09b0a7f014fbe0bc1898ee/src/schemas/json/github-workflow.json
  cue import -p github -f -l '#Workflow:' internal/ci/github/github.actions.workflow.schema.json

build-github-workflows:
  cue cmd regenerate ./internal/ci/github

run-webserver:
  #!/usr/bin/env bash
  set -euo pipefail
  set -o allexport ; source ./.env ; set +o allexport; ./tmp/webserver

go-build target:
  go build -o "tmp/{{target}}" "./cmd/{{target}}"

install-templ:
  go install github.com/a-h/templ/cmd/templ@latest

install-node-packages:
  npm install

clean:
  rm -f static/css/tw.css views/{components,pages}/*_templ.go

staticcheck: build-tw-css build-templ
  staticcheck ./...

check:
  go test -race ./...
