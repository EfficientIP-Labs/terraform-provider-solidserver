data "solidserver_ip_address" "myFirstIPAddressData" {
  depends_on = [solidserver_ip_address.myFirstIPAddress]
  name   = solidserver_ip_address.myFirstIPAddress.name
  space  = solidserver_ip_address.myFirstIPAddress.space
}