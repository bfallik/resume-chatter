
default:
  just --list

tw:
	npx tailwindcss -i input.css -o static/css/tw.css

templ:
  templ generate

air-main: templ
  go build -o tmp/main .

# PYTHONPATH is necessary because the grpc buf back-end produces an import line like:
#   from chat.v1 import chat_pb2 as chat_dot_v1_dot_chat__pb2
langchain:
  ( set -o allexport ; source ./.env ; set +o allexport; \
    cd python/src; PYTHONPATH=$(pwd)/protocgenpy pdm run python -m resume_chatter )

mypy:
  ( cd python; pdm run mypy --strict --python-executable .venv/bin/python src/resume_chatter )
