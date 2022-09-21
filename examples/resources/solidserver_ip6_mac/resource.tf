resource "solidserver_ip6_mac" "myFirstIP6MacAassoc" {
  space   = "${solidserver_ip_space.myFirstSpace.name}"
  address = "${solidserver_ip6_address.myFirstIP6Address.address}"
  mac     = "06:16:26:36:46:56"
}