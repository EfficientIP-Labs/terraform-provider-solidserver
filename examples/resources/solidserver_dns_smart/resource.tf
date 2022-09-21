resource "solidserver_dns_smart" "myFirstDnsSMART" {
  name       = "myfirstdnssmart.priv"
  arch       = "multimaster"
  comment    = "My First DNS SMART Autmatically created"
  recursion  = true
  forward    = "first"
  forwarders = ["10.0.0.42", "10.0.0.43"]
  allow_query     = ["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/24"]
  allow_recursion = ["172.16.0.0/12", "10.0.0.0/8", "192.168.0.0/24"]
}