data "solidserver_ip6_pool" "myFirstIPv6PoolData" {
  depends_on = [solidserver_ip6_subnet.myFirstIPv6Pool]
  name   = solidserver_ip6_subnet.myFirstIPv6Pool.name
  subnet = solidserver_ip6_subnet.myFirstIPv6Pool.subnet
  space  = solidserver_ip6_subnet.myFirstIPv6Pool.space
}