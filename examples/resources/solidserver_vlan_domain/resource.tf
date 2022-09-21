resource "solidserver_vlan_domain" "myFirstVxlanDomain" {
  name   = "myFirstVxlanDomain"
  vxlan  = true
  class  = "CUSTOM_VXLAN_DOMAIN"
  class_parameters = {
    LOCATION = "PARIS"
  }
}