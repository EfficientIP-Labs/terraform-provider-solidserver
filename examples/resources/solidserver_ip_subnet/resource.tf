resource "solidserver_ip_subnet" "myFirstIPBlock" {
  space            = "${solidserver_ip_space.myFirstSpace.name}"
  request_ip       = "10.0.0.0"
  prefix_size      = 8
  name             = "myFirstIPBlock"
  terminal         = false
}

resource "solidserver_ip_subnet" "myFirstIPSubnet" {
  space            = "${solidserver_ip_space.myFirstSpace.name}"
  block            = "${solidserver_ip_subnet.myFirstIPBlock.name}"
  prefix_size      = 24
  name             = "myFirstIPSubnet"
  gateway_offset   = -1
  class            = "VIRTUAL"
  class_parameters = {
    vnid = "12666"
  }
}