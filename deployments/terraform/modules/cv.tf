# container with python cv (cctv-visionfusion-cv)
resource "docker_container" "cctv-vf_cv_container" {
  name    = "cctv-visionfusion-cv"
  image   = var.cv_image 
  restart = "unless-stopped"

  networks_advanced {
    name = docker_network.cctv_vf_net.name
  }

  env = [
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "RMQ_VHOST=${var.rmq_vhost}",
    "RMQ_QUEUE_CV_FEED=${var.rmq_queue_cv_feed}",
    "CV_RMQ_USER=${var.cv_rmq_user}",
    "CV_RMQ_PASS=${var.cv_rmq_pass}",
    "SEAWEED_HOST=${var.seaweed_host}",
    "SEAWEED_API_PORT=${var.seaweed_api_port}",
    "CV_SEAWEED_USER=${var.seaweed_cv_user}",
    "CV_SEAWEED_PASS=${var.seaweed_cv_pass}",
    "CV_MODEL_PATH=${var.cv_model_path}",
    "CV_MODEL_NAMES_PATH=${var.cv_model_names_path}"
  ]

  depends_on = [
    docker_network.cctv_vf_net,
    terraform_data.rabbitmq_policy_provisioning,
    docker_container.seaweedfs
  ]
}