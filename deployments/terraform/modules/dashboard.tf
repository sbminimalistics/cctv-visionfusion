# node.js container (cctv-visionfusion-dashboard)
resource "docker_container" "cctv-vf_dashboard_container" {
  name    = "cctv-visionfusion-dashboard"
  image   = var.dashboard_image
  restart = "unless-stopped"
  network_mode = "bridge"

  ports {
    internal = 3000
    external = 3000
  }

  networks_advanced {
    name = docker_network.cctv_vf_net.name
  }

  env = [
    "DB_HOST=${var.db_host}",
    "DB_PORT=${var.db_port}",
    "DB_NAME=${var.db_name}",
    "DASHBOARD_DB_USER=${var.dashboard_db_user}",
    "DASHBOARD_DB_PASS=${var.dashboard_db_pass}",
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "DASHBOARD_DB_SCHEMA=${var.dashboard_db_schema}",
    "DASHBOARD_RMQ_USER=${var.dashboard_rmq_user}",
    "DASHBOARD_RMQ_PASS=${var.dashboard_rmq_pass}",
    "RMQ_VHOST=${var.rmq_vhost}",
    "RMQ_QUEUE_DASHBOARD_DEVICE_EVENTS_TRIGGERED=${var.rmq_queue_dashboard_device_events_triggered}",
    "RMQ_QUEUE_DASHBOARD_COMPOSITE_EVENTS_TRIGGERED=${var.rmq_queue_dashboard_composite_events_triggered}",
  ]

  depends_on = [
    docker_network.cctv_vf_net,
    terraform_data.rabbitmq_policy_provisioning,
    docker_container.seaweedfs,
    docker_container.cctv-vf_db
  ]
}