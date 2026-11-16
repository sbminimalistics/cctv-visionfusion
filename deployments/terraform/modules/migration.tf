resource "docker_container" "migration_container" {
  name     = "cctv-visionfusion-migration"
  image    = var.migration_image
  restart  = "no"
  attach   = true
  must_run = false
  network_mode = "bridge"

  networks_advanced {
    name = docker_network.cctv_vf_net.name
  }

  env = [
    "DB_HOST=${var.db_host}",
    "DB_PORT=${var.db_port}",
    "DB_NAME=${var.db_name}",
    "MIGRATION_DB_USER=${var.migration_db_user}",
    "MIGRATION_DB_PASS=${var.migration_db_pass}",
    "BACKEND_DB_SCHEMA=${var.backend_db_schema}",
    "DASHBOARD_DB_SCHEMA=${var.dashboard_db_schema}",
    "CONFIG_PATH=/app/config.yml"
  ]

  depends_on = [
    docker_network.cctv_vf_net,
    docker_container.cctv-vf_db
  ]
}

