terraform {
  required_providers {
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0.0"
    }
    rabbitmq = {
      source  = "cyrilgdn/rabbitmq"
      version = "~> 1.10" 
    }
    time = {
      source  = "hashicorp/time"
      version = "~> 0.11"
    }
  }
}

provider "docker" {
  host = "ssh://${var.ssh_user}@${var.remote_ubuntu_ip}:22"
  ssh_opts = ["-i", "'${var.ssh_key_path}'", "-o", "IdentitiesOnly=yes"]
}
