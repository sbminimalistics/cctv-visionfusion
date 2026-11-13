variable "remote_ubuntu_ip" {
  type        = string
  default     = "192.168.0.178"
}

variable "ssh_user" {
  type        = string
  default     = "ubuntu"
}

variable "ssh_key_path" {
  type        = string
  sensitive   = true
}

variable "db_name" {
  type        = string
  default     = "cctv-vf_db"
}

variable "db_port" {
  type        = string
  default     = "5432"
}

variable "admin_db_pass" {
  type        = string
  sensitive   = true
}

variable "backend_db_user" {
  type        = string
  default     = "backend_db"
}

variable "backend_db_pass" {
  type        = string
  sensitive   = true
}

variable "backend_db_schema" {
  type        = string
  default     = "cctvvf_backend"
}

variable "dashboard_db_user" {
  type        = string
  default     = "dashboard_db"
}

variable "dashboard_db_pass" {
  type        = string
  sensitive   = true
}

variable "dashboard_db_schema" {
  type        = string
  default     = "cctvvf_dashboard"
}

variable "rmq_host" {
  type        = string
  default     = "rabbitmq-broker"
}

variable "rmq_port" {
  type        = string
  default     = "5672"
}

variable "rmq_vhost" {
  type        = string
  default     = "cctv-vf"
}

variable "rmq_erlang_cookie" {
  type        = string
  sensitive   = true
}

variable "admin_rmq_pass" {
  type        = string
  sensitive   = true
}

variable "backend_rmq_user" {
  type        = string
  default     = "backend_rmq"
}

variable "backend_rmq_pass" {
  type        = string
  sensitive   = true
}

variable "dashboard_rmq_user" {
  type        = string
  default     = "dashboard_rmq"
}

variable "dashboard_rmq_pass" {
  type        = string
  sensitive   = true
}

variable "downloader_rmq_user" {
  type        = string
  default     = "downloader_rmq"
}

variable "downloader_rmq_pass" {
  type        = string
  sensitive   = true
}

variable "cv_rmq_user" {
  type        = string
  default     = "cv_rmq"
}

variable "cv_rmq_pass" {
  type        = string
  sensitive   = true
}

variable "rmq_queue_dashboard_feed" {
  type        = string
  default     = "dashboard_feed"
}

variable "rmq_queue_dashboard_triggers" {
  type        = string
  default     = "dashboard_triggers"
}

variable "rmq_queue_downloader_feed" {
  type        = string
  default     = "downloader_feed"
}

variable "rmq_queue_cv_feed" {
  type        = string
  default     = "image_refs"
}

variable "seaweed_host" {
  type        = string
  default     = "seaweedfs"
}

variable "seaweed_master_port" {
  type        = string
  default     = "9333"
}

variable "seaweed_volume_port" {
  type        = string
  default     = "8080"
}

variable "seaweed_api_port" {
  type        = string
  default     = "8333"
}

variable "seaweed_admin_user" {
  type        = string
  default     = "admin"
}

variable "seaweed_admin_pass" {
  type        = string
  sensitive   = true
}

variable "seaweed_bucket" {
  type        = string
  default     = "pictures"
}

variable "seaweed_downloader_user" {
  type        = string
  default     = "downloader_worker"
}

variable "seaweed_downloader_pass" {
  type        = string
  sensitive   = true
}

variable "seaweed_dashboard_user" {
  type        = string
  default     = "dashboard_worker"
}

variable "seaweed_dashboard_pass" {
  type        = string
  sensitive   = true
}

variable "seaweed_cv_user" {
  type        = string
  default     = "cv_worker"
}

variable "seaweed_cv_pass" {
  type        = string
  sensitive   = true
}

variable "seaweed_host_dir" {
  type        = string
  default     = null
}

variable "cv_model_path" {
  type        = string
  default     = "models/yolov10s.onnx"
}

variable "cv_model_names_path" {
  type        = string
  default     = "models/coco.names"
}

variable "downloader_backend_base_url" {
  type        = string
  default     = "http://127.0.0.1:8001"
}

variable "downloader_save_path_prefix" {
  type        = string
  default     = "unprocessed/"
}

variable "backend_config_path" {
  type        = string
  default     = "services/backend/config.yml"
}
