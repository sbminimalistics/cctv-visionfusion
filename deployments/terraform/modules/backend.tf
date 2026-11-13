resource "docker_container" "backend_container" {
  name    = "cctv-visionfusion-backend"
  image   = var.backend_image
  restart = "unless-stopped"

  upload {
    content = file("${path.module}/../../../${var.backend_config_path}")
    file    = "/app/config.yml"
  }

  ports {
    internal = 8888
    external = 8888
  }

  networks_advanced {
    name = docker_network.cctv_vf_net.name
  }

  env = [
    "DB_HOST=${var.remote_ubuntu_ip}",
    "DB_PORT=${var.db_port}",
    "DB_NAME=${var.db_name}",
    "BACKEND_DB_USER=${var.backend_db_user}",
    "BACKEND_DB_PASS=${var.backend_db_pass}",
    "BACKEND_DB_SCHEMA=${var.backend_db_schema}",
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "BACKEND_RMQ_USER=${var.backend_rmq_user}",
    "BACKEND_RMQ_PASS=${var.backend_rmq_pass}",
    "RMQ_VHOST=${var.rmq_vhost}",
    "CONFIG_PATH=/app/config.yml"
  ]

  depends_on = [
    docker_network.cctv_vf_net,
    terraform_data.rabbitmq_policy_provisioning
  ]
}

