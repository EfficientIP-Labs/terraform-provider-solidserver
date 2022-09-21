resource "solidserver_app_node" "myFirstNode" {
  name         = "myFirstNode"
  application  = "${solidserver_app_application.myFirstApplicaton.name}"
  fqdn         = "${solidserver_app_application.myFirstApplicaton.fqdn}"
  pool         = "${solidserver_app_pool.myFirstPool.name}"
  address      = "127.0.0.1"
  weight       = 1
  healthcheck  = "tcp"
  healthcheck_parameters {
    tcp_port = "443"
  }
}