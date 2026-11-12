# python cv image build
resource "terraform_data" "remote_cv_build" {
    triggers_replace = [
        filesha256("../../services/cv/requirements.txt"),
        filesha256("../../services/cv/Dockerfile"),
        filesha256("../../services/cv/main.py")
    ]

    connection {
        type        = "ssh"
        user        = var.ssh_user
        host        = var.remote_ubuntu_ip
        private_key = file(var.ssh_key_path)
    }

    provisioner "remote-exec" {
        inline = [
            # if it exists but is not a directory (e.g. a stuck file), delete it
            "[ -e /tmp/cv_build_context ] && [ ! -d /tmp/cv_build_context ] && rm -f /tmp/cv_build_context",
            # if it is a directory, clean its old contents so we start fresh
            "[ -d /tmp/cv_build_context ] && rm -rf /tmp/cv_build_context/*",
            "mkdir -p /tmp/cv_build_context"
        ]
    }

    # copy requirements and source files to the remote server's temp directory
    provisioner "file" {
        source      = "../../services/cv/models"
        destination = "/tmp/cv_build_context/models"
    }
    provisioner "file" {
        source      = "../../services/cv/proto"
        destination = "/tmp/cv_build_context/proto"
    }
    provisioner "file" {
        source      = "../../services/cv/config.py"
        destination = "/tmp/cv_build_context/config.py"
    }
    provisioner "file" {
        source      = "../../services/cv/detector.py"
        destination = "/tmp/cv_build_context/detector.py"
    }
    provisioner "file" {
        source      = "../../services/cv/Dockerfile"
        destination = "/tmp/cv_build_context/Dockerfile"
    }
    provisioner "file" {
        source      = "../../services/cv/main.py"
        destination = "/tmp/cv_build_context/main.py"
    }
    provisioner "file" {
        source      = "../../services/cv/processor.py"
        destination = "/tmp/cv_build_context/processor.py"
    }
    provisioner "file" {
        source      = "../../services/cv/rabbitmq.py"
        destination = "/tmp/cv_build_context/rabbitmq.py"
    }
    provisioner "file" {
        source      = "../../services/cv/requirements.txt"
        destination = "/tmp/cv_build_context/requirements.txt"
    }
    provisioner "file" {
        source      = "../../services/cv/seaweed.py"
        destination = "/tmp/cv_build_context/seaweed.py"
    }

    # execute the build natively on the remote machine's actual hardware
    provisioner "remote-exec" {
        inline = [
            "echo 'about to run docker build...'",
            "set -x",
            "cd /tmp/cv_build_context",
            "docker build -t cctv-visionfusion-cv:latest .",
            "rm -rf /tmp/cv_build_context"
        ]
    }
}

# container with python cv (cctv-visionfusion-cv)
resource "docker_container" "cctv-vf_cv_container" {
  name    = "cctv-visionfusion-cv"
  image   = "cctv-visionfusion-cv:latest" 
  restart = "unless-stopped"

  networks_advanced {
    name = docker_network.cctv_vf_net.name
  }

  env = [
    "RMQ_HOST=${var.rmq_host}",
    "RMQ_PORT=${var.rmq_port}",
    "RMQ_VHOST=${var.rmq_vhost}",
    "RMQ_QUEUE_CV_FEED=${var.rmq_queue_cv_feed}",
    "CV_RMQ_USER=${var.cv_rmq_user}",
    "CV_RMQ_PASS=${var.cv_rmq_pass}",
    "SEAWEED_HOST=${var.seaweed_host}",
    "SEAWEED_API_PORT=${var.seaweed_api_port}",
    "CV_SEAWEED_USER=${var.seaweed_cv_user}",
    "CV_SEAWEED_PASS=${var.seaweed_cv_pass}",
    "CV_MODEL_PATH=${var.cv_model_path}",
    "CV_MODEL_NAMES_PATH=${var.cv_model_names_path}",
  ]

  depends_on = [
    terraform_data.remote_cv_build,
    docker_network.cctv_vf_net,
    terraform_data.rabbitmq_policy_provisioning,
    docker_container.seaweedfs
  ]
}