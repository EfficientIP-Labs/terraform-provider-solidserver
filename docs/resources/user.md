---
page_title: "solidserver_user Resource - SOLIDserver"
subcategory: ""
description: |-
  User resource allows to creat and manage local SOLIDserver users who
  can connect through Web GUI and use API(s).
---

# solidserver_user (Resource)

User resource allows to creat and manage local SOLIDserver users who
can connect through Web GUI and use API(s).

## Example Usage

```terraform
resource "solidserver_user" "myFirstUser" {
   login = "jsmith"
   password = "a_very_c0mpl3x_P@ssw0rd"
   description = "My Very First User Resource"
   last_name = "Smith"
   first_name = "John"
   email = "j.smith@efficientip.com"
   groups = [ "${solidserver_usergroup.grp_admin.name}" ]
}
```
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `groups` (Set of String) The group id set for this user.
- `login` (String) The login of the user.
- `password` (String) The password of the user.

### Optional

- `class_parameters` (Map of String) The class parameters associated to the user.
- `description` (String) The description of the user.
- `email` (String) The email address of the user.
- `first_name` (String) The first name of the user.
- `last_name` (String) The last name of the user.

### Read-Only

- `id` (String) The ID of this resource.

