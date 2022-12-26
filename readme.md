# zbpack

## What is it?

`zbpack` (pronounced "Zeabur Pack") is a tool to analyze your code, determine a build plan and build a container image.

Zeabur Pack is used as the build tool for [Zeabur](https://zeabur.com/home)

## Usage

```bash
zbpack <path-to-your-code>
```

## Development

1. Fork this repo
2. Clone your fork
3. Edit the code
4. Run `mkdir bin` if it doesn't exist
5. Run `go build -o ./bin ./...`
6. Test with `./bin/zbpack <path-to-your-code>`
