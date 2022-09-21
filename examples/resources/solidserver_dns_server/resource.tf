resource "solidserver_dns_server" "myFirstDnsServer" {
  name       = "myfirstdnsserver.priv"
  address    = "127.0.0.1"
  login      = "admin"
  password   = "admin"
  forward    = "first"
  forwarders = ["10.0.0.42", "10.0.0.43"]
  allow_query     = ["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/24"]
  allow_recursion = ["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/24"]
  smart      = "${solidserver_dns_smart.myFirstDnsSMART.name}"
  smart_role = "master"
  comment    = "My First DNS Server Autmatically created"
}