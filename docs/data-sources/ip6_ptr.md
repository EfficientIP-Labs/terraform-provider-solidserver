---
page_title: "solidserver_ip6_ptr Data Source - SOLIDserver"
subcategory: ""
description: |-
  IPv6 PTR data-source allows to easily convert an IPv6 address into a DNS PTR format.
---

# solidserver_ip6_ptr (Data Source)

IPv6 PTR data-source allows to easily convert an IPv6 address into a DNS PTR format.


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `address` (String) The IPv6 address to convert into PTR domain name.

### Read-Only

- `dname` (String) The PTR record FQDN associated to the IPv6 address.
- `id` (String) The ID of this resource.

