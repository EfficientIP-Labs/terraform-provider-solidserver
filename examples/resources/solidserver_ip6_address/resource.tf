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