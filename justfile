
default:
  just --list

build-tw-css:
	npx tailwindcss -i input.css -o static/css/tw.css

build-templ:
  templ generate

build-webserver: build-tw-css build-templ (go-build "webserver")

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
