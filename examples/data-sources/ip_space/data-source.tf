data "solidserver_ip_space" "myFirstSpaceData" {
  depends_on = [solidserver_ip_space.myFirstSpace]
  name       = solidserver_ip_space.myFirstSpace.name
}