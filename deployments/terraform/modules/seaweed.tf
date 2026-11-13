locals {
  seaweed_dir = coalesce(var.seaweed_host_dir, "/home/${var.ssh_user}/s3")
}

resource "null_resource" "seaweed_remote_setup" {
  connection {
    type        = "ssh"
    user        = var.ssh_user
    host        = var.remote_ubuntu_ip
    private_key = file(var.ssh_key_path)
  }

  provisioner "remote-exec" {
    inline = [
      "mkdir -p ${local.seaweed_dir}/data",
      "mkdir -p ${local.seaweed_dir}/config",
      "chown -R 1000:1000 ${local.seaweed_dir}"
    ]
  }

  provisioner "file" {
    content = templatefile("${path.module}/s3.json.tmpl", {
      downloader_user = var.seaweed_downloader_user
      downloader_pass = var.seaweed_downloader_pass
      cv_user         = var.seaweed_cv_user
      cv_pass         = var.seaweed_cv_pass
      dashboard_user  = var.seaweed_dashboard_user
      dashboard_pass  = var.seaweed_dashboard_pass
      admin_user      = var.seaweed_admin_user
      admin_pass      = var.seaweed_admin_pass
      bucket          = var.seaweed_bucket
    })
    destination = "${local.seaweed_dir}/config/s3.json"
  }
}

resource "docker_container" "seaweedfs" {
  name    = "seaweedfs"
  image   = "chrislusf/seaweedfs:latest"
  user    = "1000:1000"
  restart = "always"

  ports {
    internal = 9333
    external = var.seaweed_master_port
  }
  ports {
    internal = 8080
    external = var.seaweed_volume_port
  }
  ports {
    internal = 8333
    external = var.seaweed_api_port
  }

  command = ["server", "-dir=/data", "-s3", "-s3.config=/etc/seaweedfs/s3.json"]

  volumes {
    host_path      = "${local.seaweed_dir}/config/s3.json"
    container_path = "/etc/seaweedfs/s3.json"
  }

  volumes {
    host_path      = "${local.seaweed_dir}/data"
    container_path = "/data"
  }

  depends_on = [null_resource.seaweed_remote_setup]
}