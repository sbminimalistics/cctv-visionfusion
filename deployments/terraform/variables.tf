variable "remote_ubuntu_ip" {
  type        = string
  description = "The IP address of your remote Ubuntu machine"
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
  default     = "vms_db"
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
  default     = "vms_backend"
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
  default     = "vms_dashboard"
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
  default     = "vms"
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

variable "admin_minio_pass" {
  type        = string
  sensitive   = true
  description = "The minIO admin password used by Terraform"
}

variable "minio_host" {
  type        = string
  description = "minIO service host"
  default     = "vms"
}

variable "minio_port" {
  type        = string
  description = "minIO service port"
  default     = "9000"
}

variable "dashboard_minio_user" {
  type        = string
  description = "Username for dashboard service"
  default     = "dashboard_minio"
}

variable "dashboard_minio_pass" {
  type        = string
  sensitive   = true
  description = "Password for dashboard service"
}

variable "downloader_minio_user" {
  type        = string
  description = "Username for downloader service"
  default     = "downloader_minio"
}

variable "downloader_minio_pass" {
  type        = string
  sensitive   = true
  description = "Password for downloader service"
}
