data "solidserver_ip6_address" "myFirstIPv6AddressData" {
  name   = solidserver_ip6_address.myFirstIPv6Address.name
  space  = solidserver_ip6_address.myFirstIPv6Address.space
}