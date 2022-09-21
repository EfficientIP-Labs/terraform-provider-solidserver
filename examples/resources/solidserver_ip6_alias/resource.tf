resource "solidserver_ip6_alias" "myFirstIP6Alias" {
  space  = "${solidserver_ip_space.myFirstSpace.name}"
  address = "${solidserver_ip6_address.myFirstIP6Address.address}"
  name   = "myfirstip6cnamealias.mycompany.priv"
}