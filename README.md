## About

*NOTE* - Currently unstable and liable to change.

GitHub Action to statically analyze container images for vulnerabilities using [Claircore](https://github.com/quay/claircore/).

___

- [About](#about)
- [Usage](#usage)
  - [Image path](#image-path)
  - [Image ref](#image-ref)
  - [Image ref with auth](#image-ref-with-auth)
- [Customizing](#customizing)
  - [inputs](#inputs)
- [Releases](#releases)

## Usage

### Image path

```yaml
name: Clair

on:
  push:
    branches:
      - 'main'
  pull_request:
    branches:
      - 'main'
jobs:
  docker-build:
    name: "Docker Build"
    runs-on: ubuntu-latest
    steps:

      - name: Checkout code
        uses: actions/checkout@v2

      - name: Build an image from Dockerfile
        run: |
          docker build -t a-really/great-app:${{ github.sha }} .

      - name: Save Docker image
        run: |
          docker save -o ${{ github.sha }} a-really/great-app:${{ github.sha }}

      - name: Run Clair V4
        uses: quay/clair-action@main
        with:
          image-path: ${{ github.sha }}
          format: sarif
          output: clair_results.sarif
  
      - name: Upload sarif
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: clair_results.sarif
```

### Image ref

```yaml
name: Clair
on:
  push:
    branches:
      - 'main'
jobs:
  docker:
    runs-on: ubuntu-latest
    steps:
      -
        name: Checkout
        uses: actions/checkout@v2
      -
        name: Set up QEMU
        uses: docker/setup-qemu-action@v2
      -
        name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      -
        name: Login to DockerHub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}
      -
        name: Build and push
        uses: docker/build-push-action@v3
        with:
          context: .
          push: true
          tags: user/app:latest
      - 
        name: Run Clair V4
        uses: quay/clair-action@main
        with:
          image-ref: user/app:latest
          format: sarif
          output: clair_results.sarif
  
      - name: Upload sarif
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: clair_results.sarif
```

### Image ref with auth

The decision was taken (that might be changeable) to just ask for the .docker/config.json file (or whereever you keep your container registry authentication configuration). The reasons for this are:
* Most people in their workflows are going to have logged into docker, or can do it easily with the docker-login action. This action already accounts for the various registry special cases.
* The clair-action container does no need to depend on the `docker` binary. 
Here is an example of how to define a workflow to use the clair-action on a private image that exists in a registry:

```yaml
name: ci

on:
  push:
    branches:
      - 'main'
  pull_request:
    branches:
      - 'main'
jobs:
  docker-pull-vulns:
    name: "Docker Pull and get vulns"
    runs-on: ubuntu-latest
    steps:
      - name: Docker login
        uses: docker/login-action@v2
        with:
          registry: quay.io
          username: ${{ secrets.QUAY_USERNAME }}
          password: ${{ secrets.QUAY_ROBOT_TOKEN }}
      - name: Copy config
        run: |
          cp ${HOME}/.docker/config.json config.json
      - name: Run Clair V4
        uses: quay/clair-action@main
        with:
          image-ref: quay.io/crozzy/quay-test:v3.4.7-15
          format: sarif
          output: clair_results.sarif
          docker-config-dir: /
      - name: Upload sarif
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: clair_results.sarif
```

## Customizing

### inputs

Following inputs can be used as `step.with` keys

| Name                | Type   | Required | default          | Description                                                                                                                                                                                |
| ------------------- | ------ | -------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `image-ref`         | String | yes*     | -                | The reference to an image in a container registry, currently this needs to be public (e.g., `quay.io/projectquay/clair:nightly`)                                                           |
| `image-path`        | String | yes*     | -                | Where on the filesystem the image was saved, i.e. the --output-flag from the `docker save` command the action require either this or `image-ref` to be defined (e.g., `/tmp/my-image.tar`) |
| `format`            | String | no       | `clair`          | The output format of the report, currently `clair`, `sarif` and `quay` are supported.                                                                                                      |
| `output`            | String | yes      | -                | The file path where the report gets saved (e.g., /tmp/my-image-report.sarif)                                                                                                               |
| `return-code`       | String | no       | `0`              | A code to return from the process if Clair found vulnerabilities. (e.g., `1`)                                                                                                              |
| `db-file-url`       | String | no       | liable to change | Optional param to specify your own url where the zstd compressed sqlite3 DB lives.                                                                                                         |
| `docker-config-dir` | String | no       | -                | Optional param to specify the docker (or other) config dir to allow for pulling of layers from private images                                                                              |


\* either `image-ref` or `image-path` need to be defined.

## Releases

Before tagging make sure to update the [Dockerfile](Dockerfile), this must happen for the action to use the correct container. The container is pre-built to keep latency as low as possible, pushing a tag should trigger that container build that is subsequently pushed to [Quay.io](https://quay.io/projectquay/clair-action).

```sh
# Update Dockerfile with new $TAG
git tag -as $TAG HEAD
git push upstream $TAG
gh workflow view release --web # if you're partial to that kind of thing
```
