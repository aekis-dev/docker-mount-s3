// docker-bake.hcl
// This file defines the build configuration for multi-platform Docker images

// Define the platforms we want to build for
variable "PLATFORMS" {
  default = ["linux/arm64", "linux/arm64"]
}

// Define all the Mountpoint S3 versions to build
variable "MOUNTPOINT_VERSIONS" {
  default = ["1.15.0", "1.16.0", "1.16.1", "1.16.2", "1.17.0", "1.18.0", "1.19.0"]
}

// Function to generate tags for each version
function "tags" {
  params = [version]
  result = [
    "aekis/docker-mount-s3:${version}"
  ]
}

// Default group that builds all versions
group "default" {
  targets = [for v in MOUNTPOINT_VERSIONS : "docker-mount-s3-${replace(v, ".", "-")}"]
}

// Dynamic target generation for each version
target "docker-mount-s3" {
  name = "docker-mount-s3-${replace(version, ".", "-")}"
  dockerfile = "src/Dockerfile"
  platforms = PLATFORMS
  tags = tags(version)
  args = {
    MOUNTPOINT_VERSION = version
  }
  matrix = {
    version = MOUNTPOINT_VERSIONS
  }
}

// Target for building the latest version with "latest" tag
target "latest" {
  inherits = ["docker-mount-s3"]
  args = {
    MOUNTPOINT_VERSION = "1.19.0"
  }
  tags = concat(
    tags("1.19.0"),
    ["aekis/docker-mount-s3:latest"]
  )
}

// Development target for local testing (builds only for current platform)
target "dev" {
  inherits = ["docker-mount-s3"]
  args = {
    MOUNTPOINT_VERSION = "1.19.0"
  }
  tags = ["aekis/docker-mount-s3:dev"]
  platforms = []  // Empty means current platform only
}