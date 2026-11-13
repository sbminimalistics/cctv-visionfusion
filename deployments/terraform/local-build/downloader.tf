resource "docker_image" "cctv-vf_downloader" {
  name = "cctv-visionfusion-downloader:latest"

  build {
    context    = "../../services/downloader"
    dockerfile = "Dockerfile"
  }

  keep_locally = true
}
