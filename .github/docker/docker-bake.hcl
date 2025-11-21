variable "IMAGE_REPO" {
  default = "ghcr.io/chaptersix"
}

variable "IMAGE_SHA_TAG" {}

variable "IMAGE_BRANCH_TAG" {}

variable "CLI_SHA" {
  default = ""
}

variable "VERSION" {
  default = "dev"
}

variable "TAG_LATEST" {
  default = false
}

# Alpine base image with digest for reproducible builds
variable "ALPINE_IMAGE" {
  default = "alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412"
}

target "cli" {
  dockerfile = ".github/docker/cli.Dockerfile"
  context = "."
  tags = compact([
    "${IMAGE_REPO}/temporal-cli:${IMAGE_SHA_TAG}",
    "${IMAGE_REPO}/temporal-cli:${VERSION}",
    TAG_LATEST ? "${IMAGE_REPO}/temporal-cli:latest" : "",
  ])
  platforms = ["linux/amd64", "linux/arm64"]
  args = {
    ALPINE_IMAGE = "${ALPINE_IMAGE}"
  }
  labels = {
    "org.opencontainers.image.title" = "temporal"
    "org.opencontainers.image.description" = "Temporal CLI"
    "org.opencontainers.image.url" = "https://github.com/temporalio/cli"
    "org.opencontainers.image.source" = "https://github.com/temporalio/cli"
    "org.opencontainers.image.licenses" = "MIT"
    "org.opencontainers.image.revision" = "${CLI_SHA}"
    "org.opencontainers.image.created" = timestamp()
  }
}
