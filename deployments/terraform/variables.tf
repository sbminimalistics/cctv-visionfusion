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

variable "db_admin_pass" {
  type        = string
  sensitive   = true
  description = "Database admin password"
}

variable "db_producer_user" {
  type        = string
  description = "Database user name used for producer service"
  default     = "producer"
}

variable "db_producer_pass" {
  type        = string
  sensitive   = true
  description = "Database password used for producer service"
}

variable "db_producer_schema" {
  type        = string
  description = "Database schema used for producer service"
  default     = "vms_producer"
}

variable "db_consumer_user" {
  type        = string
  description = "Database user name used for consumer service"
  default     = "consumer"
}

variable "db_consumer_pass" {
  type        = string
  sensitive   = true
  description = "Database password used for consumer service"
}

variable "db_consumer_schema" {
  type        = string
  description = "Database schema used for consumer service"
  default     = "vms_consumer"
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

variable "rmq_admin_pass" {
  type        = string
  sensitive   = true
  description = "The main bootstrapper admin password used by Terraform"
}

variable "rmq_producer_user" {
  type        = string
  description = "Username for any service acting as a message producer"
  default     = "rmq_producer"
}

variable "rmq_producer_pass" {
  type        = string
  sensitive   = true
  description = "Password for any service acting as a message producer"
}

variable "rmq_consumer_user" {
  type        = string
  description = "Username for consumer services"
  default     = "rmq_consumer"
}

variable "rmq_consumer_pass" {
  type        = string
  sensitive   = true
  description = "Password for consumer services"
}
