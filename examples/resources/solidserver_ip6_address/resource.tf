resource "solidserver_ip6_address" "myFirstIP6Address" {
  space   = "${solidserver_ip_space.myFirstSpace.name}"
  subnet  = "${solidserver_ip6_subnet.myFirstIP6Subnet.name}"
  name    = "myfirstip6address"
  device  = "${solidserver_device.myFirstDevice.name}"
  class   = "AWS_VPC_ADDRESS"
  class_parameters = {
    interfaceid = "eni-d5b961d5"
  }
}

// When using IPv6-MAC association, consider using the lifecycle property on the associated IPv6 address for statefull management of the MAC address.
resource "solidserver_ip6_address" "myFirstIP6Address" {
  space   = "${solidserver_ip_space.myFirstSpace.name}"
  subnet  = "${solidserver_ip6_subnet.myFirstIP6Subnet.name}"
  name    = "myfirstip6address"
  lifecycle {
    ignore_changes = ["mac"]
  }
}