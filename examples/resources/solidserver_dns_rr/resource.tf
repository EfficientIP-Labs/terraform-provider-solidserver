resource "solidserver_dns_rr" "aaRecord" {
  dnsserver = "ns.mycompany.priv"
  dnsview   = "Internal"
  dnszone   = "mycompany.priv"
  name      = "aarecord.mycompany.priv"
  type      = "A"
  value     = "127.0.0.1"
}

// In order to create a PTR, you can leverage the data-source "solidserver_ip_ptr"
// to generate the proper FQDN from an IP address
data "solidserver_ip_ptr" "myFirstIPPTR" {
  address = "${solidserver_ip_address.myFirstIPAddress.address}"
}

resource "solidserver_dns_rr" "aaRecord" {
  dnsserver = "ns.mycompany.priv"
  dnsview   = "Internal"
  dnszone   = "mycompany.priv"
  name      = "${solidserver_ip_ptr.myFirstIPPTR.dname}"
  type      = "PTR"
  value     = "myapp.mycompany.priv"
}