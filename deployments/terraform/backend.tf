# compile the Go binary locally for Linux architecture
resource "null_resource" "go_build" {
  triggers = {
    always_run = "${timestamp()}"
    # rebuild only if actual Go source files or configurations change
    dir_sha1 = sha1(join("", [for f in fileset("../../services/vms-backend-service", "**/*.go") : filesha1("../../services/vms-backend-service/${f}")]))
  }

  provisioner "local-exec" {
    command = "cd ../../services/vms-backend-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -o vms-backend-app ./cmd/vms-app/main.go"
  }
}

# use a completely isolated, standalone creation resource
resource "terraform_data" "remote_docker_build" {
  depends_on = [null_resource.go_build]

  triggers_replace = [
    timestamp()
  ]

  provisioner "remote-exec" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }

    inline = [
      "mkdir -p /tmp/vms-build"
    ]
  }

  provisioner "file" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }
    source      = "../../services/vms-backend-service/vms-backend-app"
    destination = "/tmp/vms-build/vms-backend-app"
  }

  provisioner "file" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }
    source      = "../../services/vms-backend-service/config.yml"
    destination = "/tmp/vms-build/config.yml"
  }

  provisioner "file" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }
    source      = "../../services/vms-backend-service/Dockerfile"
    destination = "/tmp/vms-build/Dockerfile"
  }

  provisioner "remote-exec" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }
    inline = [
      "cd /tmp/vms-build",
      "docker build -t vms-backend-app:latest .",
      "rm -rf /tmp/vms-build"
    ]
  }
}

# resource dedicated solely to the destroy cleanup
resource "terraform_data" "remote_docker_cleanup" {
  input = {
    ssh_user     = var.ssh_user
    remote_ip    = var.remote_ubuntu_ip
    ssh_key_path = var.ssh_key_path
  }

  provisioner "remote-exec" {
    when = destroy

    connection {
      type        = "ssh"
      user        = self.output.ssh_user
      host        = self.output.remote_ip
      private_key = file(self.output.ssh_key_path)
    }

    inline = [
      "docker rmi vms-backend-app:latest || true"
    ]
  }
}

resource "docker_container" "vms_container" {
  name    = "vms-backend"
  image   = "vms-backend-app:latest"
  restart = "unless-stopped"

  ports {
    internal = 8080
    external = 8080
  }

  networks_advanced {
    name = docker_network.vms_net.name
  }

  env = [
    "DB_HOST=${var.remote_ubuntu_ip}",
    "DB_PORT=${var.db_port}",
    "DB_NAME=${var.db_name}",
    "BACKEND_DB_USER=${var.backend_db_user}",
    "BACKEND_DB_PASS=${var.backend_db_pass}",
    "BACKEND_DB_SCHEMA=${var.backend_db_schema}",
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "BACKEND_RMQ_USER=${var.backend_rmq_user}",
    "BACKEND_RMQ_PASS=${var.backend_rmq_pass}",
    "RMQ_VHOST=${var.rmq_vhost}",
    "RMQ_EXCHANGE_BACKEND_EVENTS=${var.rmq_exchange_backend_events}",
    "RMQ_EXCHANGE_BACKEND_TRIGGERS=${var.rmq_exchange_backend_triggers}"
  ]

  depends_on = [
    terraform_data.remote_docker_build,
    terraform_data.remote_docker_cleanup,
    terraform_data.rabbitmq_policy_provisioning
  ]
}

