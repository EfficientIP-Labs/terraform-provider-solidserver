data "solidserver_ip6_subnet" "myFirstIP6SubnetData" {
  depends_on = [solidserver_ip6_subnet.myFirstIP6Subnet]
  name       = solidserver_ip6_subnet.myFirstIP6Subnet.name
  space      = solidserver_ip6_subnet.myFirstIP6Subnet.space
}