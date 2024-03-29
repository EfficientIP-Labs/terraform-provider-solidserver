---
page_title: "solidserver_ip6_pool Resource - SOLIDserver"
subcategory: ""
description: |-
  IPv6 Pool resource allows to create and manage ranges of IPv6 addresses for specific usage such as: provisioning,
  planning or migrations. IPv6 Pools can also be used to delegate one or several ranges of IPv6 addresses to groups
  of administrators or to restrict access to some users.
---

# solidserver_ip6_pool (Resource)

IPv6 Pool resource allows to create and manage ranges of IPv6 addresses for specific usage such as: provisioning,
planning or migrations. IPv6 Pools can also be used to delegate one or several ranges of IPv6 addresses to groups
of administrators or to restrict access to some users.

## Example Usage

```terraform
resource "solidserver_ip6_pool" "myFirstIPPool" {
  space            = "${solidserver_ip_space.myFirstSpace.name}"
  subnet           = "${solidserver_ip6_subnet.mySecondIP6Subnet.name}"
  name             = "myFirstIP6Pool"
  start            = "${solidserver_ip6_subnet.mySecondIP6Subnet.address}"
  size             = 2
}
```
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `end` (String) The IPv6 pool's higher IPv6 address.
- `name` (String) The name of the IPv6 pool to create.
- `space` (String) The name of the space into which creating the IPv6 pool.
- `start` (String) The IPv6 pool's lower IPv6 address.
- `subnet` (String) The name of the parent IP subnet into which creating the IPv6 pool.

### Optional

- `class` (String) The class associated to the IPv6 pool.
- `class_parameters` (Map of String) The class parameters associated to the IPv6 pool.
- `dhcp_range` (Boolean) Specify wether to create the equivalent DHCP v6 range, or not (Default: false).

### Read-Only

- `id` (String) The ID of this resource.
- `prefix` (String) The prefix of the parent subnet of the pool.
- `prefix_size` (Number) The size prefix of the parent subnet of the pool.

