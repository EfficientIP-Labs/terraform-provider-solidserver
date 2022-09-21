resource "solidserver_dns_rr" "aaRecord" {
  dnsserver = "ns.mycompany.priv"
  dnsview   = "Internal"
  dnszone   = "mycompany.priv"
  name      = "aarecord.mycompany.priv"
  type      = "A"
  value     = "127.0.0.1"
}