resource "docker_image" "backend" {
  name         = "ghcr.io/sbminimalistics/cctv-visionfusion/backend:${var.backend_image_tag}"
  keep_locally = false
}

resource "docker_image" "cv" {
  name         = "ghcr.io/sbminimalistics/cctv-visionfusion/cv:${var.cv_image_tag}"
  keep_locally = false
}

resource "docker_image" "dashboard" {
  name         = "ghcr.io/sbminimalistics/cctv-visionfusion/dashboard:${var.dashboard_image_tag}"
  keep_locally = false
}

resource "docker_image" "downloader" {
  name         = "ghcr.io/sbminimalistics/cctv-visionfusion/downloader:${var.downloader_image_tag}"
  keep_locally = false
}

module "stack" {
  source = "../modules"

  backend_image = docker_image.backend.image_id
  cv_image = docker_image.cv.image_id
  dashboard_image = docker_image.dashboard.image_id
  downloader_image = docker_image.downloader.image_id

  remote_ubuntu_ip = var.remote_ubuntu_ip
  ssh_user = var.ssh_user
  ssh_key_path = var.ssh_key_path

  db_name = var.db_name
  db_port = var.db_port
  admin_db_pass = var.admin_db_pass
  backend_db_user = var.backend_db_user
  backend_db_pass = var.backend_db_pass
  backend_db_schema = var.backend_db_schema
  dashboard_db_user = var.dashboard_db_user
  dashboard_db_pass = var.dashboard_db_pass
  dashboard_db_schema = var.dashboard_db_schema

  rmq_host = var.rmq_host
  rmq_port = var.rmq_port
  rmq_vhost = var.rmq_vhost
  rmq_erlang_cookie = var.rmq_erlang_cookie
  admin_rmq_pass = var.admin_rmq_pass
  backend_rmq_user = var.backend_rmq_user
  backend_rmq_pass = var.backend_rmq_pass
  downloader_rmq_user = var.downloader_rmq_user
  downloader_rmq_pass = var.downloader_rmq_pass
  dashboard_rmq_user = var.dashboard_rmq_user
  dashboard_rmq_pass = var.dashboard_rmq_pass
  cv_rmq_user = var.cv_rmq_user
  cv_rmq_pass = var.cv_rmq_pass
  rmq_queue_dashboard_feed = var.rmq_queue_dashboard_feed
  rmq_queue_dashboard_triggers = var.rmq_queue_dashboard_triggers
  rmq_queue_downloader_feed = var.rmq_queue_downloader_feed
  rmq_queue_cv_feed = var.rmq_queue_cv_feed

  seaweed_downloader_user = var.seaweed_downloader_user
  seaweed_downloader_pass = var.seaweed_downloader_pass
  seaweed_cv_user = var.seaweed_cv_user
  seaweed_cv_pass = var.seaweed_cv_pass
  seaweed_dashboard_user = var.seaweed_dashboard_user
  seaweed_dashboard_pass = var.seaweed_dashboard_pass
  seaweed_admin_user = var.seaweed_admin_user
  seaweed_admin_pass = var.seaweed_admin_pass
  seaweed_bucket = var.seaweed_bucket
  seaweed_host = var.seaweed_host
  seaweed_host_dir = var.seaweed_host_dir
  seaweed_master_port = var.seaweed_master_port
  seaweed_volume_port = var.seaweed_volume_port
  seaweed_api_port = var.seaweed_api_port

  cv_model_path = var.cv_model_path
  cv_model_names_path = var.cv_model_names_path

  downloader_backend_base_url = var.downloader_backend_base_url
  downloader_save_path_prefix = var.downloader_save_path_prefix

  backend_config_path = var.backend_config_path
}