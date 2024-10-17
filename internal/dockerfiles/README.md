# Dockerfiles

This directory contains the Dockerfiles used to build the Docker images for a project according to the build plan.

Planners can select the Dockerfile and the stage to use. For example,

```toml
zeaburImage = deno
```

selects the [`selectable/deno.Dockerfile`](selectable/deno.Dockerfile) to build the Docker image.

For serverless cases, you may want to build more than one variant of the Docker image, such as the *containerized* version and the *serverless* version. In this case, you can use `zeaburImage` and `zeaburImageStage` to specify the Dockerfile and the stage. For example,

```toml
zeaburImage = dart-flutter
zeaburImageStage = target-static
```

selects the [`selectable/dart-flutter.Dockerfile`](selectable/dart-flutter.Dockerfile) and the `target-static` stage to build the Docker image.

The `ARG` in the Dockerfile corresponds to the field in the build plan. For example, if there is a build plan like:

```toml
build = "flutter build web"
zeaburImage = dart-flutter
zeaburImageStage = target-static
```

and there is an `ARG` in the Dockerfile:

```dockerfile
ARG build
RUN $build
```

then the `build` field in the build plan will be passed to the Dockerfile, resulting in:

```dockerfile
RUN flutter build web
```
