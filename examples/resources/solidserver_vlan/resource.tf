resource "solidserver_vlan" "myFirstVxlan" {
  vlan_domain      = "${solidserver_vlan_domain.myFirstVxlanDomain.name}"
  name             = "myFirstVxlan"
}