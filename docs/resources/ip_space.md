---
page_title: "solidserver_ip_space Resource - SOLIDserver"
subcategory: ""
description: |-
  Space resource allows to create and manage the highest level objets in the SOLIDserver's IPAM module
  organization, the entry point of any IPv4 or IPv6 addressing plan. Spaces allow to manage unique ranges
  of IP addresses.
---

# solidserver_ip_space (Resource)

Space resource allows to create and manage the highest level objets in the SOLIDserver's IPAM module
organization, the entry point of any IPv4 or IPv6 addressing plan. Spaces allow to manage unique ranges
of IP addresses.

## Example Usage

```terraform
resource "solidserver_ip_space" "myFirstSpace" {
  name   = "myFirstSpace"
  class  = "CUSTOM_SPACE"
  class_parameters = {
    LOCATION = "PARIS"
  }
}
```
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the IP space to create.

### Optional

- `class` (String) The class associated to the IP space.
- `class_parameters` (Map of String) The class parameters associated to IP space.

### Read-Only

- `id` (String) The ID of this resource.

