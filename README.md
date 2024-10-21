# zbpack

`zbpack` (pronounced “Zeabur Pack”) aims to automatically analyze the language, version, and framework used based on the source code and package the service into the most suitable deployment form, such as static resources, cloud functions, containers, or multiple types by one click. It is mainly used in the [Zeabur](https://zeabur.com) platform as the build tool to determine the code type and build the container image automatically (without writing the Dockerfile manually.)

## Components

`zbpack` is consist of the following components:

- Planners: The planners are responsible for analyzing the source code and determining the type of the code. The planners are located in the `internal` directory.
- `zeaburpack` library: The `zeaburpack` library is the library of `zbpack` that can be used in your platform directly. It is located in the `pkg/zeaburpack` directory.
- CLI: The CLI is the command-line interface of `zbpack` for testing purpose. It is located in the `cmd/zbpack` directory.

## Usage

### Common Part

1. Fork the repository and clone it to your local machine.
2. Make sure you have Go installed on your machine. You can download it from the official website: <https://golang.org/dl/>
3. Navigate to the root of the project and run `go mod download` to download the necessary dependencies.
4. Make sure [buildctl](https://github.com/moby/buildkit) is installed and buildkitd is running.

### `zbpack`

`zbpack` analyzes your projects, constructs the image recipes, and builds the container image automatically.

1. Build the binary by running `go build -o zbpack ./cmd/zbpack/main.go`.
2. Run the binary with `./zbpack [the directory to analyze and build]`.

You should see the `build plan` block and the subsequent `build log` block. The `build plan` block shows the metadata and the information (“recipes”) to build an image of this project. The `build log` block shows the build log of the container image, which is outputted by `docker build`.

Use `-i` or `--info` to show the build plan only.

```bash
$ ./zbpack --info corepack-project

╔══════════════════════════ Build Plan ═════════════════════════╗
║ provider         │ nodejs                                     ║
║───────────────────────────────────────────────────────────────║
║ startCmd         │ node index.js                              ║
║───────────────────────────────────────────────────────────────║
║ packageManager   │ pnpm                                       ║
║───────────────────────────────────────────────────────────────║
║ framework        │ none                                       ║
║───────────────────────────────────────────────────────────────║
║ nodeVersion      │ 16                                         ║
║───────────────────────────────────────────────────────────────║
║ installCmd       │ pnpm install                               ║
╚═══════════════════════════════════════════════════════════════╝
```

Get some more usage information by using `-h` or `--help`.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for more information.

## License

`zbpack` is licensed under the [Mozilla Public License 2.0](./LICENSE). According to MPL 2.0 (and the summarization of Bing AI), if you want to use, modify or distribute MPL 2.0 software, you have the following rights and obligations:

- You have the right to use, modify and distribute MPL 2.0 software for any purpose, without paying any fees.
- You have the obligation to preserve the license notice, copyright notice and disclaimer in MPL 2.0 source code files.
- You have the obligation to disclose any modifications you make to MPL 2.0 source code files, and to provide them to others under MPL 2.0 or a more permissive license.
- You have the right to combine MPL 2.0 source code files with source code files under other licenses in a software, but you cannot change the license of MPL 2.0 source code files.
- You have the right to choose to provide MPL 2.0 source code files to others under another compatible Copyleft license, such as GNU GPL, LGPL or AGPL.

## Contributors

<p align="center">
<a href="https://github.com/zeabur/zbpack/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=zeabur/zbpack" />
</a></p>
