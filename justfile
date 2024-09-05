
default:
  just --list

tw:
	npx tailwindcss -i input.css -o static/css/tw.css

templ:
  templ generate

air-main: templ
  go build -o tmp/main .

langchain:
  ( set -o allexport ; source ./.env ; set +o allexport; cd python/src; pdm run python -m resume_chatter.main )