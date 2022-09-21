data "solidserver_dns_view" "DnsViewData" {
  dnsserver = "ns.local"
  name = "testview"
}