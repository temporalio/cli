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

variable "CLI_SHA" {
  default = ""
}

variable "VERSION" {
  default = "dev"
}

target "cli" {
  dockerfile = "Dockerfile"
  context = "."
  tags = [
    "${IMAGE_NAMESPACE}/${IMAGE_NAME}:${IMAGE_SHA_TAG}",
    "${IMAGE_NAMESPACE}/${IMAGE_NAME}:${VERSION}",
  ]
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
