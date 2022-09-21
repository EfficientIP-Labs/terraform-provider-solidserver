resource "solidserver_app_application" "myFirstApplicaton" {
  name         = "MyFirsApp"
  fqdn         = "myfirstapp.priv"
  gslb_members = ["ns0.priv", "ns1.priv"]
  class        = "INTERNAL_APP"
  class_parameters = {
    owner = "MR. Smith"
    contact = "a.smith@mycompany.priv"
  }
}