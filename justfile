
default:
  just --list

build-tw-css:
	npx tailwindcss -i input.css -o static/css/tw.css

build-templ:
  templ generate

build-webserver: build-tw-css build-templ (go-build "webserver")

build-github-workflows-import-schema:
  # NOTE: this version is the only one that currently works
  curl -o internal/ci/github/github.actions.workflow.schema.json https://raw.githubusercontent.com/SchemaStore/schemastore/5ffe36662a8fcab3c32e8fbca39c5253809e6913/src/schemas/json/github-workflow.json
  # NOTE: but the schema field needs to be replaced
  contents="$( jq '.["$schema"] = "https://json-schema.org/draft/2020-12/schema"' internal/ci/github/github.actions.workflow.schema.json )" && \
  echo -E "${contents}" > internal/ci/github/github.actions.workflow.schema.json
  cue import -p github -f -l '#Workflow:' internal/ci/github/github.actions.workflow.schema.json

build-github-workflows:
  cue cmd regenerate ./internal/ci/github

run-webserver:
  #!/usr/bin/env bash
  set -euo pipefail
  set -o allexport ; source ./.env ; set +o allexport; ./tmp/webserver

go-build target:
  go build -o tmp/{{target}} ./cmd/{{target}}

install-templ:
  go install github.com/a-h/templ/cmd/templ@latest

install-node-packages:
  npm install

clean:
  rm -f static/css/tw.css views/{components,pages}/*_templ.go

staticcheck: build-tw-css build-templ
  staticcheck ./...
