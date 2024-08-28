# resume-chatter
a chat bot for resumes


## Development

Based on docs from [goship.it](https://goship.it/get-started).

Install [templ](https://templ.guide/), [npm](https://docs.npmjs.com/about-npm), and [just](https://just.systems/) dependencies:

```
> go install github.com/a-h/templ/cmd/templ@latest
...
> sudo dnf install nodejs-npm
...
```

Then install TailwindCSS and DaisyUI:

```
> npm i -D tailwindcss @tailwindcss/typography daisyui
```

Ensure `templ` is on your `PATH` (needed for vscode and the `just local` target):

```
> which templ                                                         ✔ 
~/go/bin/templ
```

Invoke `just tw` and `just local` in two terminals to run the local dev server and automatically regenerate targets for changes source files.
