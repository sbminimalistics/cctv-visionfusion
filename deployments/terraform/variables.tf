variable "remote_ubuntu_ip" {
  type        = string
  description = "remote machine IP"
  default     = "192.168.0.178"
}

variable "ssh_user" {
  type        = string
  default     = "ubuntu"
}

variable "ssh_key_path" {
  type        = string
  description = "Path to the local SSH private key file"
  sensitive   = true
}

variable "db_name" {
  type        = string
  description = "Database name"
  default     = "cctv-vf_db"
}

variable "db_port" {
  type        = string
  description = "Database name"
  default     = "5432"
}

variable "admin_db_pass" {
  type        = string
  sensitive   = true
  description = "Database admin password"
}

variable "backend_db_user" {
  type        = string
  description = "Database user name used for backend service"
  default     = "backend_db"
}

variable "backend_db_pass" {
  type        = string
  sensitive   = true
  description = "Database password used for backend service"
}

variable "backend_db_schema" {
  type        = string
  description = "Database schema used for producer service"
  default     = "cctvvf_backend"
}

variable "dashboard_db_user" {
  type        = string
  description = "Database user name used for dashboard service"
  default     = "dashboard_db"
}

variable "dashboard_db_pass" {
  type        = string
  sensitive   = true
  description = "Database password used for dashboard service"
}

variable "dashboard_db_schema" {
  type        = string
  description = "Database schema used for dashboard service"
  default     = "cctvvf_dashboard"
}

variable "rmq_host" {
  type        = string
  description = "RabbitMQ host"
  default     = "rabbitmq-broker"
}

variable "rmq_port" {
  type        = string
  description = "RabbitMQ port"
  default     = "5672"
}

variable "rmq_vhost" {
  type        = string
  description = "RabbitMQ vhost"
  default     = "cctv-vf"
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
  default     = "backend_rmq"
}

variable "backend_rmq_pass" {
  type        = string
  sensitive   = true
  description = "Password for backend service"
}

variable "dashboard_rmq_user" {
  type        = string
  description = "Username for dashboard service"
  default     = "dashboard_rmq"
}

variable "dashboard_rmq_pass" {
  type        = string
  sensitive   = true
  description = "Password for dashboard service"
}

variable "downloader_rmq_user" {
  type        = string
  description = "Username for downloader service"
  default     = "downloader_rmq"
}

variable "downloader_rmq_pass" {
  type        = string
  sensitive   = true
  description = "Password for downloader service"
}

variable "rmq_queue_dashboard_feed" {
  type        = string
  default     = "dashboard_feed"
}

variable "rmq_queue_dashboard_triggers" {
  type        = string
  default     = "dashboard_triggers"
}

variable "seaweed_host" {
  type        = string
  description = "SeaweedFS host"
  default     = "seaweedfs"
}

variable "seaweed_master_port" {
  type        = string
  description = "SeaweedFS master service port"
  default     = "9333"
}

variable "seaweed_volume_port" {
  type        = string
  description = "SeaweedFS volume service port"
  default     = "8080"
}

variable "seaweed_api_port" {
  type        = string
  description = "SeaweedFS api service port"
  default     = "8333"
}

variable "seaweed_admin_user" {
  type        = string
  description = "SeaweedFS admin user name"
  default     = "admin"
}

variable "seaweed_admin_pass" {
  type        = string
  sensitive   = true
  description = "SeaweedFS admin password"
}

variable "seaweed_bucket" {
  type        = string
  description = "SeaweedFS s3 bucket name"
  default     = "pictures"
}

variable "seaweed_downloader_user" {
  type        = string
  description = "Username for downloader service"
  default     = "downloader_worker"
}

variable "seaweed_downloader_pass" {
  type        = string
  sensitive   = true
  description = "Password for downloader service"
}

variable "seaweed_dashboard_user" {
  type        = string
  description = "Username for dashboard service"
  default     = "dashboard_worker"
}

variable "seaweed_dashboard_pass" {
  type        = string
  sensitive   = true
  description = "Password for dashboard service"
}

variable "seaweed_cv_user" {
  type        = string
  description = "Username for cv service"
  default     = "cv_worker"
}

variable "seaweed_cv_pass" {
  type        = string
  sensitive   = true
  description = "Password for cv service"
}
