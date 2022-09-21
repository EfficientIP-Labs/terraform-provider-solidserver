resource "solidserver_dns_forward_zone" "myFirstForwardZone" {
  dnsserver = "ns.priv"
  name       = "fwd.mycompany.priv"
  forward    = "first"
  forwarders = ["10.10.8.8", "10.10.4.4"]
}