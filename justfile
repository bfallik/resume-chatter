
tw:
	@npx tailwindcss -i input.css -o static/css/tw.css --watch

local:
    #!/usr/bin/env bash
    set -euxo pipefail
    templ generate -watch -proxy="http://localhost:8080" -open-browser=false -cmd="go run main.go"
