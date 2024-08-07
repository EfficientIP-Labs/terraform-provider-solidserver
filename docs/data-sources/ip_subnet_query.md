---
page_title: "solidserver_ip_subnet_query Data Source - SOLIDserver"
subcategory: ""
description: |-
  IP subnet query data-source allows to retrieve information about the first IPv4 subnet matching given criterias, including its meta-data.
---

# solidserver_ip_subnet_query (Data Source)

IP subnet query data-source allows to retrieve information about the first IPv4 subnet matching given criterias, including its meta-data.

## Example Usage

```terraform
data "solidserver_ip_subnet_query" "mySecondIPSubnetQueriedData" {
  depends_on       = [solidserver_ip_subnet.mySecondIPSubnet]
  query            = "tag_network_vnid = '12666' AND subnet_allocated_percent < '90.0'"
  tags             = "network.vnid"
}
```
<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `query` (String) The query used to find the first matching subnet.

### Optional

- `orderby` (String) The query used to find the first matching subnet.
- `tags` (String) The tags to be used to find the first matching subnet in the query.

### Read-Only

- `address` (String) The IP subnet address.
- `class` (String) The class associated to the IP subnet.
- `class_parameters` (Map of String) The class parameters associated to IP subnet.
- `gateway` (String) The subnet's computed gateway.
- `id` (String) The ID of this resource.
- `name` (String) The name of the IP subnet.
- `netmask` (String) The IP subnet netmask.
- `prefix` (String) The IP subnet prefix.
- `prefix_size` (Number) The IP subnet's prefix length (ex: 24 for a '/24').
- `space` (String) The space associated to the IP subnet.
- `terminal` (Boolean) The terminal property of the IP subnet.
- `vlan_domain` (String) The optional vlan Domain associated with the subnet.
- `vlan_id` (Number) The optional vlan ID associated with the subnet.
- `vlan_name` (String) The optional vlan Name associated with the subnet.
- `vlan_range` (String) The optional vlan Range associated with the subnet.

