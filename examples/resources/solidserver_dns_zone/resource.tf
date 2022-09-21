resource "solidserver_dns_zone" "myFirstZone" {
  dnsserver = "ns.priv"
  name      = "mycompany.priv"
  type      = "master"
  space     = "${solidserver_ip_space.myFirstSpace.name}"
  createptr = false
}