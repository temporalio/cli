variable "IMAGE_REPO" {
  default = "ghcr.io"
}

variable "IMAGE_NAMESPACE" {
  default = ""
}

variable "IMAGE_NAME" {
  default = "temporal"
}

variable "GITHUB_REPOSITORY" {
  default = "temporalio/cli"
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



target "cli" {
  dockerfile = "Dockerfile"
  context = "."
  tags = compact([
    IMAGE_REPO == "" ? "${IMAGE_NAMESPACE}/${IMAGE_NAME}:${IMAGE_SHA_TAG}" : "${IMAGE_REPO}/${IMAGE_NAMESPACE}/${IMAGE_NAME}:${IMAGE_SHA_TAG}",
    IMAGE_REPO == "" ? "${IMAGE_NAMESPACE}/${IMAGE_NAME}:${VERSION}" : "${IMAGE_REPO}/${IMAGE_NAMESPACE}/${IMAGE_NAME}:${VERSION}",
    TAG_LATEST ? (IMAGE_REPO == "" ? "${IMAGE_NAMESPACE}/${IMAGE_NAME}:latest" : "${IMAGE_REPO}/${IMAGE_NAMESPACE}/${IMAGE_NAME}:latest") : "",
  ])
  platforms = ["linux/amd64", "linux/arm64"]
  labels = {
    "org.opencontainers.image.title" = "temporal"
    "org.opencontainers.image.description" = "Temporal CLI"
    "org.opencontainers.image.url" = "https://github.com/${GITHUB_REPOSITORY}"
    "org.opencontainers.image.source" = "https://github.com/${GITHUB_REPOSITORY}"
    "org.opencontainers.image.licenses" = "MIT"
    "org.opencontainers.image.revision" = "${CLI_SHA}"
    "org.opencontainers.image.created" = timestamp()
  }
}
