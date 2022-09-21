resource "solidserver_ip6_pool" "myFirstIPPool" {
  space            = "${solidserver_ip_space.myFirstSpace.name}"
  subnet           = "${solidserver_ip6_subnet.mySecondIP6Subnet.name}"
  name             = "myFirstIP6Pool"
  start            = "${solidserver_ip6_subnet.mySecondIP6Subnet.address}"
  size             = 2
}