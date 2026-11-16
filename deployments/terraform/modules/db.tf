locals {
  db_dir = coalesce(var.db_host_dir, "/home/${var.ssh_user}/db")
}

resource "null_resource" "db_remote_setup" {
  connection {
    type        = "ssh"
    user        = var.ssh_user
    host        = var.remote_ubuntu_ip
    private_key = file(var.ssh_key_path)
    timeout     = "30s"
  }

  provisioner "remote-exec" {
    inline = [
      "mkdir -p ${local.db_dir}/data",
      "mkdir -p ${local.db_dir}/config",
      "docker run --rm -v ${local.db_dir}:/target alpine chown -R $(id -u):$(id -g) /target/config"
    ]
  }

  provisioner "file" {
    content = templatefile("${path.module}/postgres-init-users.sql.tmpl", {
      backendUser     = var.backend_db_user
      backendPass     = var.backend_db_pass
      backendSchema   = var.backend_db_schema
      dashboardUser   = var.dashboard_db_user
      dashboardPass   = var.dashboard_db_pass
      dashboardSchema = var.dashboard_db_schema
      updaterUser     = var.migration_db_user
      updaterPass     = var.migration_db_pass
    })
    destination = "${local.db_dir}/config/init-users.sql"
  }

  # hand ownership to Postgres (UID 70) after upload finishes
  provisioner "remote-exec" {
    inline = [
      "docker run --rm -v ${local.db_dir}:/target alpine sh -c 'chown -R 70:70 /target && chmod 700 /target/data && chmod 644 /target/config/init-users.sql'"
    ]
  }
}

resource "docker_container" "cctv-vf_db" {
  name    = "cctv-visionfusion-db"
  image   = "postgres:16-alpine"
  restart = "unless-stopped"
  network_mode = "bridge"

  networks_advanced {
    name = docker_network.cctv_vf_net.name
  }

  env = [
    "POSTGRES_DB=${var.db_name}",
    "POSTGRES_USER=admin",
    "POSTGRES_PASSWORD=${var.admin_db_pass}"
  ]

  volumes {
    host_path      = "${local.db_dir}/config/init-users.sql"
    container_path = "/docker-entrypoint-initdb.d/init-multi-user.sql"
    read_only      = true
  }

  volumes {
    host_path      = "${local.db_dir}/data"
    container_path = "/var/lib/postgresql/data"
  }

  depends_on = [null_resource.db_remote_setup]
}