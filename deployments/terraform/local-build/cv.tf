# python cv image build
resource "terraform_data" "remote_cv_build" {
    triggers_replace = [
        filesha256("../../../services/cv/requirements.txt"),
        filesha256("../../../services/cv/Dockerfile"),
        filesha256("../../../services/cv/main.py")
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
        source      = "../../../services/cv/models"
        destination = "/tmp/cv_build_context/models"
    }
    provisioner "file" {
        source      = "../../../services/cv/proto"
        destination = "/tmp/cv_build_context/proto"
    }
    provisioner "file" {
        source      = "../../../services/cv/config.py"
        destination = "/tmp/cv_build_context/config.py"
    }
    provisioner "file" {
        source      = "../../../services/cv/detector.py"
        destination = "/tmp/cv_build_context/detector.py"
    }
    provisioner "file" {
        source      = "../../../services/cv/Dockerfile"
        destination = "/tmp/cv_build_context/Dockerfile"
    }
    provisioner "file" {
        source      = "../../../services/cv/main.py"
        destination = "/tmp/cv_build_context/main.py"
    }
    provisioner "file" {
        source      = "../../../services/cv/processor.py"
        destination = "/tmp/cv_build_context/processor.py"
    }
    provisioner "file" {
        source      = "../../../services/cv/rabbitmq.py"
        destination = "/tmp/cv_build_context/rabbitmq.py"
    }
    provisioner "file" {
        source      = "../../../services/cv/requirements.txt"
        destination = "/tmp/cv_build_context/requirements.txt"
    }
    provisioner "file" {
        source      = "../../../services/cv/seaweed.py"
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
