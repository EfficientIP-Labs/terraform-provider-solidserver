data "solidserver_ip_subnet_query" "mySecondIPSubnetQueriedData" {
  depends_on       = [solidserver_ip_subnet.mySecondIPSubnet]
  query            = "tag_network_vnid = '12666' AND subnet_allocated_percent < '90.0'"
  tags             = "network.vnid"
}