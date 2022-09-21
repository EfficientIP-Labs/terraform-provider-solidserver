data "solidserver_ip6_subnet_query" "mySecondIPv6SubnetQueriedData" {
  depends_on       = [solidserver_ip6_subnet.mySecondIPv6Subnet]
  query            = "tag_network_vnid = '12666' AND subnet_allocated_percent < '90.0'"
  tags             = "network.vnid"
}