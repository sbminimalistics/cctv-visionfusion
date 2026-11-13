# private network bridge
resource "docker_network" "cctv_vf_net" {
  name = "cctv-visionfusion-internal"
}
