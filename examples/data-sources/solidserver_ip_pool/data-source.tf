data "solidserver_ip_pool" "myFirstIPPoolData" {
  depends_on = [solidserver_ip_subnet.myFirstIPPool]
  name   = solidserver_ip_subnet.myFirstIPPool.name
  subnet = solidserver_ip_subnet.myFirstIPPool.subnet
  space  = solidserver_ip_subnet.myFirstIPPool.space
}