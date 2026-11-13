# node.js container (cctv-visionfusion-downloader)
resource "docker_container" "cctv-vf_downloader_container" {
  name    = "cctv-visionfusion-downloader"
  image   = var.downloader_image
  restart = "unless-stopped"

  networks_advanced {
    name = docker_network.cctv_vf_net.name
  }

  env = [
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "DOWNLOADER_RMQ_USER=${var.dashboard_rmq_user}",
    "DOWNLOADER_RMQ_PASS=${var.dashboard_rmq_pass}",
    "RMQ_VHOST=${var.rmq_vhost}",
    "SEAWEED_HOST=${var.seaweed_host}",
    "SEAWEED_API_PORT=${var.seaweed_api_port}",
    "SEAWEED_BUCKET=${var.seaweed_bucket}",
    "DOWNLOADER_SEAWEED_USER=${var.seaweed_downloader_user}",
    "DOWNLOADER_SEAWEED_PASS=${var.seaweed_downloader_pass}",

    "RMQ_QUEUE_DOWNLOADER_FEED=${var.rmq_queue_downloader_feed}",
    "DOWNLOADER_BACKEND_BASE_URL=${var.downloader_backend_base_url}",
    "DOWNLOADER_SAVE_PATH_PREFIX=${var.downloader_save_path_prefix}"
  ]

  depends_on = [
    docker_network.cctv_vf_net,
    terraform_data.rabbitmq_policy_provisioning,
    docker_container.seaweedfs
  ]
}