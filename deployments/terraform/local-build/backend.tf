locals {
  backend_src_dir = "${path.module}/../../../services/backend"

  backend_watched_files = setunion(
    fileset(local.backend_src_dir, "{Dockerfile,go.mod,go.sum}"),
    fileset(local.backend_src_dir, "**/*.go")
  )
}

resource "docker_image" "cctv-vf_backend" {
  triggers = {
    dir_sha1 = sha1(join("", [
      for f in sort(local.backend_watched_files) :
      filesha256("${local.backend_src_dir}/${f}")
    ]))
  }

  name = "cctv-visionfusion-backend:latest"

  build {
    context    = local.backend_src_dir
    dockerfile = "Dockerfile"
  }

  keep_locally = true
}
