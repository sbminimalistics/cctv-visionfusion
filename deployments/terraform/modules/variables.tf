variable "backend_image" {
  type        = string
  description = "backend image id"
  default     = null
}

variable "cv_image" {
  type        = string
  description = "cv image id"
}

variable "dashboard_image" {
  type        = string
  description = "dashboard image id"
  default     = null
}

variable "downloader_image" {
  type        = string
  description = "downloader image id"
  default     = null
}

variable "remote_ubuntu_ip" {
  type        = string
  description = "remote machine IP"
}

variable "ssh_user" {
  type        = string
}

variable "ssh_key_path" {
  type        = string
  description = "Path to the local SSH private key file"
  sensitive   = true
}

variable "db_name" {
  type        = string
  description = "Database name"
  default     = null
}

variable "db_port" {
  type        = string
  description = "Database name"
  default     = null
}

variable "admin_db_pass" {
  type        = string
  sensitive   = true
  description = "Database admin password"
}

variable "backend_db_user" {
  type        = string
  description = "Database user name used for backend service"
  default     = null
}

variable "backend_db_pass" {
  type        = string
  sensitive   = true
  description = "Database password used for backend service"
}

variable "backend_db_schema" {
  type        = string
  description = "Database schema used for producer service"
}

variable "dashboard_db_user" {
  type        = string
  description = "Database user name used for dashboard service"
}

variable "dashboard_db_pass" {
  type        = string
  sensitive   = true
  description = "Database password used for dashboard service"
  default     = null
}

variable "dashboard_db_schema" {
  type        = string
  description = "Database schema used for dashboard service"
}

variable "rmq_host" {
  type        = string
  description = "RabbitMQ host"
}

variable "rmq_port" {
  type        = string
  description = "RabbitMQ port"
}

variable "rmq_vhost" {
  type        = string
  description = "RabbitMQ vhost"
}

variable "rmq_erlang_cookie" {
  type        = string
  description = "The secret Erlang cookie for administration access"
  sensitive   = true
}

variable "admin_rmq_pass" {
  type        = string
  sensitive   = true
  description = "The main bootstrapper admin password used by Terraform"
}

variable "backend_rmq_user" {
  type        = string
  description = "Username for backend service"
}

variable "backend_rmq_pass" {
  type        = string
  sensitive   = true
  description = "Password for backend service"
}

variable "dashboard_rmq_user" {
  type        = string
  description = "Username for dashboard service"
}

variable "dashboard_rmq_pass" {
  type        = string
  sensitive   = true
  description = "Password for dashboard service"
}

variable "downloader_rmq_user" {
  type        = string
  description = "Username for downloader service"
}

variable "downloader_rmq_pass" {
  type        = string
  sensitive   = true
  description = "Password for downloader service"
}

variable "cv_rmq_user" {
  type        = string
  description = "Username for cv service"
}

variable "cv_rmq_pass" {
  type        = string
  sensitive   = true
  description = "Password for cv service"
}

variable "rmq_queue_dashboard_feed" {
  type        = string
}

variable "rmq_queue_dashboard_triggers" {
  type        = string
  default     = null
}

variable "rmq_queue_downloader_feed" {
  type        = string
}

variable "rmq_queue_cv_feed" {
  type        = string
}

variable "seaweed_host" {
  type        = string
  description = "SeaweedFS host"
}

variable "seaweed_master_port" {
  type        = string
  description = "SeaweedFS master service port"
}

variable "seaweed_volume_port" {
  type        = string
  description = "SeaweedFS volume service port"
}

variable "seaweed_api_port" {
  type        = string
  description = "SeaweedFS api service port"
}

variable "seaweed_admin_user" {
  type        = string
  description = "SeaweedFS admin user name"
}

variable "seaweed_admin_pass" {
  type        = string
  sensitive   = true
  description = "SeaweedFS admin password"
}

variable "seaweed_bucket" {
  type        = string
  description = "SeaweedFS s3 bucket name"
}

variable "seaweed_downloader_user" {
  type        = string
  description = "Username for downloader service"
}

variable "seaweed_downloader_pass" {
  type        = string
  sensitive   = true
  description = "Password for downloader service"
}

variable "seaweed_dashboard_user" {
  type        = string
  description = "Username for dashboard service"
}

variable "seaweed_dashboard_pass" {
  type        = string
  sensitive   = true
  description = "Password for dashboard service"
}

variable "seaweed_cv_user" {
  type        = string
  description = "Username for cv service"
}

variable "seaweed_cv_pass" {
  type        = string
  sensitive   = true
  description = "Password for cv service"
}

variable "seaweed_host_dir" {
  type        = string
  description = "SeaweedFS data and config dir"
}

variable "cv_model_path" {
  type        = string
  description = "Path to object detection model file"
}

variable "cv_model_names_path" {
  type        = string
  description = "Path to object detection names file"
}

variable "downloader_backend_base_url" {
  type        = string
  description = "Path to backend service api"
}

variable "downloader_save_path_prefix" {
  type        = string
  description = "Prefix for every saved object"
}

variable "backend_config_path" {
  type        = string
  description = "backend config file path within container"
}