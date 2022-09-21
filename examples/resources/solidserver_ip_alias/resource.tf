resource "solidserver_ip_alias" "myFirstIPAlias" {
  space  = "${solidserver_ip_space.myFirstSpace.name}"
  address = "${solidserver_ip_address.myFirstIPAddress.address}"
  name   = "myfirstipcnamealias.mycompany.priv"
}