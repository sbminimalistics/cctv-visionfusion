resource "null_resource" "minio_directory_setup" {
  provisioner "remote-exec" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }

    inline = [
      "mkdir -p /home/ovd/minio/data",
      "chown -R 1000:1000 /home/ovd/minio/data"
    ]
  }
}

resource "docker_container" "minio_server" {
  name    = "minio_server"
  image   = "alpine/minio:RELEASE.2025-10-15T17-29-55Z"
  command = ["server", "/data", "--console-address", ":9001"]
  user    = "1000:1000"

  ports {
    internal = 9000
    external = 9000
  }
  ports {
    internal = 9001
    external = 9001
  }

  env = [
    "MINIO_ROOT_USER=admin-super-user",
    "MINIO_ROOT_PASSWORD=super-secret-admin-password"
  ]

  volumes {
    host_path      = "/home/ovd/minio/data"
    container_path = "/data"
  }

  depends_on = [null_resource.minio_directory_setup]
}

resource "time_sleep" "wait_for_minio_boot" {
  depends_on      = [docker_container.minio_server]
  create_duration = "7s" # Gives the remote engine ample time to bind ports
}

provider "minio" {
  # FIX: Change 'localhost' to your remote server IP variable
  minio_server   = "${var.remote_ubuntu_ip}:9000" 
  minio_user     = "admin-super-user"
  minio_password = "super-secret-admin-password"
  minio_ssl      = false
}

resource "minio_s3_bucket" "images_bucket" {
  bucket     = "raw-images"
  acl        = "private"
  depends_on = [time_sleep.wait_for_minio_boot]
}

# minIO users

resource "minio_iam_user" "go_downloader" {
  name       = "go_downloader_user"
  secret     = "go_password_123" # <── FIXED: Changed from secret_key to secret
  depends_on = [time_sleep.wait_for_minio_boot]
}

resource "minio_iam_user_policy_attachment" "go_write_policy" {
  user_name   = minio_iam_user.go_downloader.name
  policy_name = "writeonly"
}

resource "minio_iam_user" "python_worker" {
  name       = "python_worker_user"
  secret     = "python_password_123" # <── FIXED: Changed from secret_key to secret
  depends_on = [time_sleep.wait_for_minio_boot]
}

resource "minio_iam_user_policy_attachment" "python_read_policy" {
  user_name   = minio_iam_user.python_worker.name
  policy_name = "readonly"
}