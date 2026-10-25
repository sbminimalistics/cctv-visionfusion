# node.js ui image build
resource "docker_image" "cctv-vf_dashboard" {
  name = "cctv-visionfusion-dashboard:latest"

  build {
    context    = "../../services/dashboard"
    dockerfile = "Dockerfile"
  }

  keep_locally = true
}

# node.js container (cctv-visionfusion-dashboard)
resource "docker_container" "cctv-vf_ui_container" {
  name    = "cctv-visionfusion-dashboard"
  image   = docker_image.cctv-vf_dashboard.image_id 
  restart = "unless-stopped"

  ports {
    internal = 3000
    external = 3000
  }

  networks_advanced {
    name = docker_network.cctv_vf_net.name
  }

  env = [
    "DB_HOST=${var.remote_ubuntu_ip}",
    "DB_PORT=5432",
    "DB_NAME=cctv-vf_production",
    "DASHBOARD_DB_USER=${var.dashboard_db_user}",
    "DASHBOARD_DB_PASS=${var.dashboard_db_pass}",
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "DASHBOARD_RMQ_USER=${var.dashboard_rmq_user}",
    "DASHBOARD_RMQ_PASS=${var.dashboard_rmq_pass}",
    "RMQ_VHOST=${var.rmq_vhost}",
    "MINIO_HOST=${var.minio_host}",
    "MINIO_PORT=${var.minio_port}",
    "MINIO_USER=${var.dashboard_minio_user}",
    "MINIO_PASS=${var.dashboard_minio_pass}",
  ]

  depends_on = [
    docker_network.cctv_vf_net,
    terraform_data.rabbitmq_policy_provisioning
  ]
}