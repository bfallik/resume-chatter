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
