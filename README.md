## About

*NOTE* - Currently unstable and liable to change.

GitHub Action to statically analyze container images for vulnerabilities using [Claircore](https://github.com/quay/claircore/).

___

* [Usage](#usage)
  * [Image path](#image-path)
  * [Image ref](#image-ref)
* [Customizing](#customizing)
  * [inputs](#inputs)
* [Troubleshooting](#troubleshooting)

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
        uses: crozzy/clair-action@main
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
        uses: crozzy/clair-action@main
        with:
          image-ref: user/app:latest
          format: sarif
          output: clair_results.sarif
  
      - name: Upload sarif
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: clair_results.sarif
```

## Customizing

### inputs

Following inputs can be used as `step.with` keys

| Name          | Type   | Required | default          | Description                                                                                                                                                                                |
|---------------|--------|----------|------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `image-ref`   | String | yes*     | -                | The reference to an image in a container registry, currently this needs to be public (e.g., `quay.io/projectquay/clair:nightly`)                                                           |
| `image-path`  | String | yes*     | -                | Where on the filesystem the image was saved, i.e. the --output-flag from the `docker save` command the action require either this or `image-ref` to be defined (e.g., `/tmp/my-image.tar`) |
| `format`      | String | no       | `clair`            | The output format of the report currently `clair` and `sarif` are supported.                                                                                                               |
| `output`      | String | yes      | -                | The file path where the report gets saved (e.g., /tmp/my-image-report.sarif)                                                                                                               |
| `return-code` | String | no       | `0`              | A code to return from the process if Clair found vulnerabilities. (e.g., `1`)                                                                                                              |
| `db-file-url` | String | no       | liable to change | Optional param to specify your own url where the zstd compressed sqlite3 DB lives.                                                                                                         |


\* either `image-ref` or `image-path` need to be defined.
