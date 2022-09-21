resource "solidserver_ip6_subnet" "myFirstIP6Block" {
  space            = "${solidserver_ip_space.myFirstSpace.name}"
  request_ip       = "2a00:2381:126d:0:0:0:0:0"
  prefix_size      = 48
  name             = "myFirstIP6Block"
  terminal         = false
}

resource "solidserver_ip6_subnet" "myFirstIP6Subnet" {
  space            = "${solidserver_ip_space.myFirstSpace.name}"
  block            = "${solidserver_ip6_subnet.myFirstIP6Block.name}"
  prefix_size      = 64
  name             = "myFirstIP6Subnet"
  gateway_offset   = 1
  class            = "VIRTUAL"
  class_parameters = {
    vnid = "12666"
  }
}