resource "solidserver_ip_space" "myFirstSpace" {
  name   = "myFirstSpace"
  class  = "CUSTOM_SPACE"
  class_parameters = {
    LOCATION = "PARIS"
  }
}