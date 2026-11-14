resource "docker_image" "cctv-vf_backend" {
  name = "cctv-visionfusion-backend:latest"

  build {
    context    = "../../../services/backend"
    dockerfile = "Dockerfile"
  }

  keep_locally = true
}
