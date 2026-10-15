# private network bridge
resource "docker_network" "vms_net" {
  name = "vms-internal"
}

# compile the Go binary locally for Linux architecture
resource "null_resource" "go_build" {
  triggers = {
    always_run = "${timestamp()}"
    # rebuild only if actual Go source files or configurations change
    dir_sha1 = sha1(join("", [for f in fileset("../../vms-backend-service", "**/*.go") : filesha1("../../vms-backend-service/${f}")]))
  }

  provisioner "local-exec" {
    command = "cd ../../vms-backend-service && set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -o vms-app ./cmd/vms-app/main.go"
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
    source      = "../../vms-backend-service/vms-app"
    destination = "/tmp/vms-build/vms-app"
  }

  provisioner "file" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }
    source      = "../../vms-backend-service/config.yml"
    destination = "/tmp/vms-build/config.yml"
  }

  provisioner "file" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }
    source      = "../../vms-backend-service/Dockerfile"
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
      "docker build -t vms-app:latest .",
      "rm -rf /tmp/vms-build"
    ]
  }
}

# resource dedicated solely to the destroy cleanup
# prevents compilation errors because it never contains any 'var.' references
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
      "docker rmi vms-app:latest || true"
    ]
  }
}

resource "docker_container" "vms_container" {
  name    = "vms-backend"
  image   = "vms-app:latest"
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
    "PRODUCER_DB_USER=${var.db_producer_user}",
    "PRODUCER_DB_PASSWORD=${var.db_producer_pass}",
    "PRODUCER_DB_SCHEMA=${var.db_producer_schema}",
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "RMQ_PRODUCER_USER=${var.rmq_producer_user}",
    "RMQ_PRODUCER_PASS=${var.rmq_producer_pass}",
    "RMQ_PRODUCER_VHOST=${var.rmq_vhost}"
  ]

  depends_on = [
    terraform_data.remote_docker_build,
    terraform_data.remote_docker_cleanup,
    terraform_data.rabbitmq_policy_provisioning
  ]
}

resource "time_sleep" "wait_for_rabbitmq" {
  depends_on = [docker_container.rabbitmq]

  create_duration = "10s"
}

# RabbitMQ

resource "docker_container" "rabbitmq" {
  name  = "rabbitmq-broker"
  image = "rabbitmq:4-management"

  ports {
    internal = var.rmq_port
    external = var.rmq_port
  }

  upload {
    content        = file("${path.module}/../../rabbitmq/rabbitmq.conf")
    file           = "/etc/rabbitmq/rabbitmq.conf"
  }

  upload {
    content        = file("${path.module}/../../rabbitmq/definitions.json")
    file           = "/etc/rabbitmq/definitions.json"
  }

  env = [
    "HOME=/tmp",
    "RABBITMQ_ERLANG_COOKIE=${var.rmq_erlang_cookie}",
    "RABBITMQ_CONFIG_FILE=/etc/rabbitmq/rabbitmq.conf",
    "RABBITMQ_DEFAULT_USER=admin",
    "RABBITMQ_DEFAULT_PASS=${var.rmq_admin_pass}"
  ]

  networks_advanced {
    name = docker_network.vms_net.name
  }

  healthcheck {
    test         = ["CMD", "rabbitmq-diagnostics", "-q", "check_running"]
    interval     = "5s"
    timeout      = "2s"
    retries      = 10
    start_period = "5s"
  }
}

resource "terraform_data" "rabbitmq_health_gate" {
  triggers_replace = [
    docker_container.rabbitmq.id
  ]

  provisioner "remote-exec" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }

    inline = [
      "echo 'Waiting for Docker engine to report container health validation...'",
      # wait until health check returns "healthy"
      "docker container wait rabbitmq-broker --condition=healthy",
      "echo 'Broker is fully verified operational!'"
    ]
  }
}

resource "terraform_data" "rabbitmq_provisioning" {
  depends_on = [terraform_data.rabbitmq_health_gate]

  triggers_replace = [
    docker_container.rabbitmq.id
  ]

  provisioner "remote-exec" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }

    inline = [
      "docker exec rabbitmq-broker rabbitmqctl wait /var/lib/rabbitmq/mnesia/rabbit.pid",

      "sleep 2",
      
      # "docker exec rabbitmq-broker rabbitmqctl add_vhost vms || true",
      
      "docker exec rabbitmq-broker rabbitmqctl delete_user \"${var.rmq_producer_user}\" || true",
      "docker exec rabbitmq-broker rabbitmqctl add_user \"${var.rmq_producer_user}\" \"${var.rmq_producer_pass}\"",
      "docker exec rabbitmq-broker rabbitmqctl set_user_tags \"${var.rmq_producer_user}\" management",
      "docker exec rabbitmq-broker rabbitmqctl set_permissions -p vms \"${var.rmq_producer_user}\" \".*\" \".*\" \".*\"",

      "docker exec rabbitmq-broker rabbitmqctl delete_user \"${var.rmq_consumer_user}\" || true",
      "docker exec rabbitmq-broker rabbitmqctl add_user \"${var.rmq_consumer_user}\" \"${var.rmq_consumer_pass}\"",
      "docker exec rabbitmq-broker rabbitmqctl set_permissions -p vms \"${var.rmq_consumer_user}\" \"^\\$\" \"^\\$\" \".*\""
    ]
  }
}

resource "terraform_data" "rabbitmq_policy_provisioning" {
  depends_on = [terraform_data.rabbitmq_provisioning]

  triggers_replace = [
    docker_container.rabbitmq.id
  ]

  provisioner "remote-exec" {
    connection {
      type        = "ssh"
      user        = var.ssh_user
      host        = var.remote_ubuntu_ip
      private_key = file(var.ssh_key_path)
    }

    inline = [
      "docker exec rabbitmq-broker rabbitmqctl set_policy -p vms camera_queue_limits \"^camera\\.\" \"{\\\"max-length\\\":5000,\\\"overflow\\\":\\\"drop-head\\\",\\\"message-ttl\\\":86400000}\" --apply-to queues"
    ]
  }
}

# node.js ui image build
resource "docker_image" "vms_ui" {
  name = "vms-ui:latest"

  build {
    context    = "../../vms-ui-service"
    dockerfile = "Dockerfile"
  }

  keep_locally = true
}

# node.js container (vms-ui)
resource "docker_container" "vms_ui_container" {
  name    = "vms-ui"
  image   = docker_image.vms_ui.image_id 
  restart = "unless-stopped"

  ports {
    internal = 3000
    external = 3000
  }

  networks_advanced {
    name = docker_network.vms_net.name
  }

  env = [
    "DB_HOST=${var.remote_ubuntu_ip}",
    "DB_PORT=5432",
    "DB_NAME=vms_production",
    "NODE_DB_USER=${var.db_consumer_user}",
    "NODE_DB_PASSWORD=${var.db_consumer_pass}",
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "RMQ_CONSUMER_USER=${var.rmq_consumer_user}",
    "RMQ_CONSUMER_PASS=${var.rmq_consumer_pass}",
    "RMQ_VHOST=vms"
  ]

  depends_on = [
    terraform_data.rabbitmq_policy_provisioning
  ]
}
