
default:
  just --list

tw:
	npx tailwindcss -i input.css -o static/css/tw.css

templ:
  templ generate

build-webserver: tw templ (go-build "webserver")

run-webserver:
  #!/usr/bin/env bash
  set -euo pipefail
  set -o allexport ; source ./.env ; set +o allexport; ./tmp/webserver

go-build target:
  go build -o tmp/{{target}} ./cmd/{{target}}

clean:
  rm -f static/css/tw.css views/{components,pages}/*_templ.go
