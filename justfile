
default:
  just --list

tw:
	npx tailwindcss -i input.css -o static/css/tw.css

templ:
  templ generate

air-main: tw templ
  go build -o tmp/main .
