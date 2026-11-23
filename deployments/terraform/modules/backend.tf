resource "docker_container" "backend_container" {
  name    = "cctv-visionfusion-backend"
  image   = var.backend_image
  restart = "unless-stopped"
  network_mode = "bridge"

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
    "DB_HOST=${var.db_host}",
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
    "CONFIG_PATH=/app/config.yml",
    "BACKEND_API_PORT=${var.backend_api_port}",
    "BACKEND_DB_LOG_LEVEL=${var.backend_db_log_level}",
    "BACKEND_LOG_LEVEL=${var.backend_log_level}",
    "RMQ_QUEUE_BACKEND_DETECTIONS_COMPLETED=${var.rmq_queue_backend_detections_completed}"
  ]

  depends_on = [
    docker_network.cctv_vf_net,
    terraform_data.rabbitmq_policy_provisioning,
    docker_container.cctv-vf_db
  ]
}

