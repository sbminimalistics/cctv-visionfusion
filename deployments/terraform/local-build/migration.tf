resource "docker_image" "cctv-vf_migration" {
  name = "cctv-visionfusion-migration:latest"

  build {
    context    = "../../../services/migration"
    dockerfile = "Dockerfile"
  }

  keep_locally = true
}
