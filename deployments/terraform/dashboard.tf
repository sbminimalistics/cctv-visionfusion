# node.js ui image build
resource "docker_image" "vms_dashboard" {
  name = "vms-dashboard:latest"

  build {
    context    = "../../services/vms-dashboard-service"
    dockerfile = "Dockerfile"
  }

  keep_locally = true
}

# node.js container (vms-dashboard)
resource "docker_container" "vms_ui_container" {
  name    = "vms-dashboard"
  image   = docker_image.vms_dashboard.image_id 
  restart = "unless-stopped"

  ports {
    internal = 3000
    external = 3000
  }

  networks_advanced {
    name = docker_network.vms_net.name
  }

  env = [
    "DB_HOST=${var.remote_ubuntu_ip}",
    "DB_PORT=5432",
    "DB_NAME=vms_production",
    "DB_USER=${var.dashboard_db_user}",
    "DB_PASSWORD=${var.dashboard_db_pass}",
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "RMQ_USER=${var.dashboard_rmq_user}",
    "RMQ_PASS=${var.dashboard_rmq_pass}",
    "RMQ_VHOST=vms",
    "MINIO_HOST=${var.minio_host}",
    "MINIO_PORT=${var.minio_port}",
    "MINIO_USER=${var.dashboard_minio_user}",
    "MINIO_PASS=${var.dashboard_minio_pass}",
  ]

  depends_on = [
    terraform_data.rabbitmq_policy_provisioning
  ]
}