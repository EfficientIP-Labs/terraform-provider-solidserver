resource "solidserver_ip_pool" "myFirstIPPool" {
  space            = "${solidserver_ip_space.myFirstSpace.name}"
  subnet           = "${solidserver_ip_subnet.mySecondIPSubnet.name}"
  name             = "myFirstIPPool"
  start            = "${solidserver_ip_subnet.mySecondIPSubnet.address}"
  size             = 2
}