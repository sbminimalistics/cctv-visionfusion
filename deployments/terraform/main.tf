# private network bridge
resource "docker_network" "vms_net" {
  name = "vms-internal"
}
