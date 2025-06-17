target "docker-metadata-action" {}

variable "APP" {
  default = "homelab-assistant"
}

variable "VERSION" {
  # Default to CalVer if not provided
  default = formatdate("YYYY.M.D", timestamp())
}

variable "SOURCE" {
  default = "https://github.com/rafaribe/homelab-assistant"
}

group "default" {
  targets = ["image-local"]
}

target "image" {
  inherits = ["docker-metadata-action"]
  args = {
    VERSION = "${VERSION}"
    BUILD_DATE = formatdate("YYYY-MM-DD'T'hh:mm:ssZ", timestamp())
  }
  labels = {
    "org.opencontainers.image.source" = "${SOURCE}"
    "org.opencontainers.image.version" = "${VERSION}"
    "org.opencontainers.image.created" = formatdate("YYYY-MM-DD'T'hh:mm:ssZ", timestamp())
  }
}

target "image-local" {
  inherits = ["image"]
  output = ["type=docker"]
  tags = ["${APP}:${VERSION}", "${APP}:latest"]
}

target "image-all" {
  inherits = ["image"]
  platforms = [
    "linux/amd64",
    "linux/arm64"
  ]
}
