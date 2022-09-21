resource "solidserver_dns_view" "myFirstDnsView" {
  depends_on      = [solidserver_dns_server.myFirstDnsServer]
  name            = "myfirstdnsview"
  dnsserver       = solidserver_dns_server.myFirstDnsServer.name
  recursion       = true
  forward         = "first"
  forwarders      = ["8.8.8.8", "8.8.4.4"]
  match_clients   = ["172.16.0.0/12", "192.168.0.0/24"]
}