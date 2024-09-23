# resume-chatter
a chat bot for resumes


## Development

Based on docs from [goship.it](https://goship.it/get-started).

Install dependencies:
* [templ](https://templ.guide/)
* [npm](https://docs.npmjs.com/about-npm)
* [just](https://just.systems/)
* [air](https://github.com/air-verse/air)

```
> go install github.com/a-h/templ/cmd/templ@latest
...
> go install github.com/air-verse/air@latest
...
> sudo dnf install nodejs-npm just
...
```

Then install TailwindCSS and DaisyUI:

```
> npm i -D tailwindcss @tailwindcss/typography daisyui
```

Ensure `templ` and `air` are on your `PATH`:

```
> which templ
~/go/bin/templ
> which air
~/go/bin/air

```

Invoke `air` to run the local dev server and automatically rebuild and reload when source files change.

```
> air

  __    _   ___
 / /\  | | | |_)
/_/--\ |_| |_| \_ v1.52.3, built with Go go1.23.0

watching .
watching cmd
watching node_modules
...
watching static
watching static/css
!exclude tmp
watching views
watching views/components
building...

Rebuilding...

🌼   daisyUI 4.12.10
├─ ✔︎ 1 theme added		https://daisyui.com/docs/themes
╰─ ❤︎ Support daisyUI project:	https://opencollective.com/daisyui


Done in 487ms.
templ generate
(✓) Complete [ updates=1 duration=7.789752ms ]
go build -o tmp/main .
running...

```

OpenAI API Notes
----------------
langchain expects an OPENAI_API_KEY env variable. The `justfile` expects this to live in a `.env` at the root of the repo.


go-plugin notes
---------------
Install [buf](https://github.com/bufbuild), as recommended in the [go-plugin example](https://github.com/hashicorp/go-plugin/blob/main/examples/grpc/proto/kv.proto):

```
> github.com/bufbuild/buf/cmd/buf@v1.40.0
```

The proto files lives in `protoc/` subdirectory. `buf` configuration lives in `buf.yml` and `buf.gen.yml`.

Basic hello world testing:

```
> go run cmd/bufserver/main.go &
...
listening on localhost:8081
> buf curl --schema . --data '{"question": "what is happening?"}' http://localhost:8081/chat.v1.ChatService/Ask
2024/09/06 11:55:30 Got a request to answer what is happening?
{}
```
