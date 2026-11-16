locals {
  downloader_src_dir = "${path.module}/../../../services/downloader"

  downloader_watched_files = setunion(
    fileset(local.downloader_src_dir, "{Dockerfile,go.mod,go.sum}"),
    fileset(local.downloader_src_dir, "**/*.go")
  )
}

resource "docker_image" "cctv-vf_downloader" {
  name = "cctv-visionfusion-downloader:latest"

  triggers = {
    dir_sha1 = sha1(join("", [
      for f in sort(local.downloader_watched_files) :
      filesha256("${local.downloader_src_dir}/${f}")
    ]))
  }

  build {
    context    = local.downloader_src_dir
    dockerfile = "Dockerfile"
  }

  keep_locally = true
}
