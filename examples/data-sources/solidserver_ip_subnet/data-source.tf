data "solidserver_ip_subnet" "myFirstIPSubnetData" {
  depends_on = [solidserver_ip_subnet.myFirstIPSubnet]
  name       = solidserver_ip_subnet.myFirstIPSubnet.name
  space      = solidserver_ip_subnet.myFirstIPSubnet.space
}