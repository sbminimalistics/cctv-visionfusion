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
    "RABBITMQ_DEFAULT_PASS=${var.admin_rmq_pass}"
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
      
      "docker exec rabbitmq-broker rabbitmqctl delete_user \"${var.backend_rmq_user}\" || true",
      "docker exec rabbitmq-broker rabbitmqctl add_user \"${var.backend_rmq_user}\" \"${var.backend_rmq_pass}\"",
      "docker exec rabbitmq-broker rabbitmqctl set_user_tags \"${var.backend_rmq_user}\" management",
      "docker exec rabbitmq-broker rabbitmqctl set_permissions -p vms \"${var.backend_rmq_user}\" \".*\" \".*\" \".*\"",

      "docker exec rabbitmq-broker rabbitmqctl delete_user \"${var.downloader_rmq_user}\" || true",
      "docker exec rabbitmq-broker rabbitmqctl add_user \"${var.downloader_rmq_user}\" \"${var.downloader_rmq_pass}\"",
      "docker exec rabbitmq-broker rabbitmqctl set_permissions -p vms \"${var.downloader_rmq_user}\" \"^\\$\" \".*\" \".*\"",

      "docker exec rabbitmq-broker rabbitmqctl delete_user \"${var.dashboard_rmq_user}\" || true",
      "docker exec rabbitmq-broker rabbitmqctl add_user \"${var.dashboard_rmq_user}\" \"${var.dashboard_rmq_pass}\"",
      "docker exec rabbitmq-broker rabbitmqctl set_permissions -p vms \"${var.dashboard_rmq_user}\" \"^\\$\" \"^\\$\" \".*\""
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