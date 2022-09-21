resource "solidserver_ip_address" "myFirstIPAddress" {
  space   = "${solidserver_ip_space.myFirstSpace.name}"
  subnet  = "${solidserver_ip_subnet.myFirstIPSubnet.name}"
  name    = "myfirstipaddress"
  device  = "${solidserver_device.myFirstDevice.name}"
  class   = "AWS_VPC_ADDRESS"
  class_parameters = {
    interfaceid = "eni-d5b961d5"
  }
}

// When using IP-MAC association, consider using the lifecycle property on the associated IP address for statefull management of the MAC address.
resource "solidserver_ip_address" "myFirstIPAddress" {
  space   = "${solidserver_ip_space.myFirstSpace.name}"
  subnet  = "${solidserver_ip_subnet.myFirstIPSubnet.name}"
  name    = "myfirstipaddress"
  lifecycle {
    ignore_changes = ["mac"]
  }
}