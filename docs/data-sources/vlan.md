---
page_title: "solidserver_vlan Data Source - SOLIDserver"
subcategory: ""
description: |-
  VLAN data-source allows to retrieve information about vlans, including meta-data.
---

# solidserver_vlan (Data Source)

VLAN data-source allows to retrieve information about vlans, including meta-data.


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the vlan.
- `vlan_domain` (String) The name of the vlan Domain.

### Optional

- `vlan_range` (String) The name of the vlan Range.

### Read-Only

- `class` (String) The class associated to the vlan.
- `class_parameters` (Map of String) The class parameters associated to vlan.
- `id` (String) The ID of this resource.
- `vlan_id` (Number) The vlan ID.

