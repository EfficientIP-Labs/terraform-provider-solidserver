resource "solidserver_app_pool" "myFirstPool" {
  name         = "myFirstPool"
  application  = "${solidserver_app_application.myFirstApplicaton.name}"
  fqdn         = "${solidserver_app_application.myFirstApplicaton.fqdn}"
  lb_mode      = latency
  affinity     = true
  affinity_session_duration = 300
}