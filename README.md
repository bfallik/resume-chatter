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


Python Setup Notes
------------------

This is an attempt to set up a modern Python dev environment based on https://www.stuartellis.name/articles/python-modern-practices/. Python is set up within the `python/` subdirectory using [pyenv](https://github.com/pyenv/pyenv) and [pdm](https://pdm-project.org/).


```
> curl https://pyenv.run | bash
```
then [setup shell environment](https://github.com/pyenv/pyenv?tab=readme-ov-file#set-up-your-shell-environment-for-pyenv).


Fedora needed some dependencies in order to build Python (from https://stribny.name/blog/install-python-dev/ and https://stackoverflow.com/questions/5459444/tkinter-python-may-not-be-configured-for-tk):

```
> sudo dnf install zlib-devel bzip2 bzip2-devel readline-devel sqlite sqlite-devel openssl-devel xz xz-devel libffi-devel findutils -y
...
> sudo dnf install tk-devel -y
...
> pyenv install 3.12
...
> pyenv local
...
> pyenv version
3.12.5 (set by /home/bfallik/sandbox/resume-chatter/.python-version)
> python --version
Python 3.12.5
```

Install [pipx](https://github.com/pypa/pipx) and add to `$PATH` and then `pdm`:

```
> pip install pipx
...
> pipx install pdm
...
> pdm info
PDM version:
  2.18.1
Python Interpreter:
  /home/bfallik/sandbox/resume-chatter/python/.venv/bin/python (3.12)
Project Root:
  /home/bfallik/sandbox/resume-chatter/python
Local Packages:
```

Lastly install [mypy](https://mypy-lang.org/):
```
> pipx install mypy
```


OpenAI API Notes
----------------
langchain expects an OPENAI_API_KEY env variable. The `justfile` expects this to live in a `.env` at the root of the repo.