# Contributing to `zbpack`

## Getting Started

See [README.md](./README.md).

## Tests

In `zbpack`, writing tests is important to ensure the quality and correctness of the code. It can catch bugs early in the development process and prevent regressions when making changes to the codebase.

There are two kinds of tests in `zbpack`:

- Unit Test and Integration Test: The example of such tests can be seen in the [`internal/nodejs/node_test.go`](internal/nodejs/node_test.go) file.
- System Test: See [tests](./tests) directory. Usually it is a demo project for checking if `zbpack` works as intended.

Once you have written your tests, you should run them before committing your code. For unit tests, you can do this by running the `go test` command in the directory containing your test files, and this command will run all the tests in your package and report any errors. For system tests, you can check if it works by running `./zbpack [folder path]` manually.

## Code Style

- **Write the tests** for every new feature you add.
- Run the tests by running `go test ./...`.
- Format your code by running `gofumpt -w .`. You may need to [install gofumpt](https://github.com/mvdan/gofumpt) before running this command.
- Lint your code before committing by running `golangci-lint run`. You may need to [install golangci-lint](https://golangci-lint.run/) before running this command.

## Commit Messages

We use the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) format for our commit messages. It makes the commit message clear for readability.

Each commit message should have a type, a scope, and a subject. Here is an example:

```plain
feat(planner/rust): Speed up the compilation

This commit speeds up the compilation by using
`cargo-chef`. It caches the Docker layer and can
speed up the compilation by 100%!
```

The scope can be:

- `cli`: The command-line interface.
- `lib`: The library exposed to the users. (`pkg/zeaburpack`)
- `planner`: The build planner. (`internal/*`)
- `utils`: The utility functions. (`internal/utils`)
- `lint`: The configuration of linters, formatters, `.editerconfig`, etc.
- (feel free to add your own scope if these can not fulfill your changes)

You can contain subscopes in your scope. For example, `cli/zbpack`.

## Pull Requests

1. Create a new branch for your changes.
2. Make your changes and commit them with clear commit messages following the guidelines above.
3. Push your branch to your fork of the repository.
4. Open a pull request against the `main` branch of the original repository.
5. Wait for a maintainer to review and merge your changes.

Thank you for your contributions!
