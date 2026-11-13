# node.js ui image build
resource "docker_image" "cctv-vf_dashboard" {
  name = "cctv-visionfusion-dashboard:latest"

  build {
    context    = "../../../services/dashboard"
    dockerfile = "Dockerfile"
  }

  keep_locally = true
}
