resource "solidserver_user" "myFirstUser" {
   login = "jsmith"
   password = "a_very_c0mpl3x_P@ssw0rd"
   description = "My Very First User Resource"
   last_name = "Smith"
   first_name = "John"
   email = "j.smith@efficientip.com"
   groups = [ "${solidserver_usergroup.grp_admin.name}" ]
}