data "solidserver_dns_zone" "myFirstDnsZoneData" {
  name = solidserver_dns_zone.myFirstZone.name
}