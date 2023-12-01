package solidserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/url"
	"strconv"
)

func dataSourceip6subnet() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceip6subnetRead,

		Description: heredoc.Doc(`
			IPv6 subnet data-source allows to retrieve information about reserved IPv6 subnets, including meta-data.
			IPv6 Subnet are key to organize the IP space, they can be blocks or subnets. Blocks reflect assigned IP
			ranges (RFC1918 or public prefixes). Subnets reflect the internal sub-division of your network.
		`),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the IPv6 subnet.",
				Required:    true,
			},
			"space": {
				Type:        schema.TypeString,
				Description: "The space associated to the IPv6 subnet.",
				Required:    true,
			},
			"address": {
				Type:        schema.TypeString,
				Description: "The IP subnet address.",
				Computed:    true,
			},
			"prefix": {
				Type:        schema.TypeString,
				Description: "The IPv6 subnet prefix.",
				Computed:    true,
			},
			"prefix_size": {
				Type:        schema.TypeInt,
				Description: "The IPv6 subnet's prefix length (ex: 64 for a '/64').",
				Computed:    true,
			},
			"terminal": {
				Type:        schema.TypeBool,
				Description: "The terminal property of the IPv6 subnet.",
				Computed:    true,
			},
			"vlan_id": {
				Type:        schema.TypeInt,
				Description: "The optional vlan ID associated with the subnet.",
				Computed:    true,
			},
			"gateway": {
				Type:        schema.TypeString,
				Description: "The  IPv6 subnet's computed gateway.",
				Computed:    true,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the IPv6 subnet.",
				Computed:    true,
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to IPv6 subnet.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceip6subnetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)
	d.SetId("")

	// Building parameters
	parameters := url.Values{}
	whereClause := "subnet6_name LIKE '" + d.Get("name").(string) + "'" +
		" and site_name LIKE '" + d.Get("space").(string) + "'"

	parameters.Add("WHERE", whereClause)

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip6_block6_subnet6_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.SetId(buf[0]["subnet6_id"].(string))

			address := hexip6toip6(buf[0]["start_ip6_addr"].(string))
			prefix_size, _ := strconv.Atoi(buf[0]["subnet6_prefix"].(string))

			d.Set("name", buf[0]["subnet6_name"].(string))
			d.Set("address", address)
			d.Set("prefix", address+"/"+buf[0]["subnet6_prefix"].(string))
			d.Set("prefix_size", prefix_size)

			if buf[0]["is_terminal"].(string) == "1" {
				d.Set("terminal", true)
			} else {
				d.Set("terminal", false)
			}

			if vlanID, vlanIDExist := buf[0]["vlmvlan_vlan_id"]; vlanIDExist {
				d.Set("vlan_id", vlanID)
			}

			d.Set("class", buf[0]["subnet6_class_name"].(string))

			// Setting local class_parameters
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["subnet6_class_parameters"].(string))
			computedClassParameters := map[string]string{}

			if gateway, gatewayExist := retrievedClassParameters["gateway"]; gatewayExist {
				d.Set("gateway", gateway[0])
			}

			for ck := range retrievedClassParameters {
				if ck != "gateway" {
					computedClassParameters[ck] = retrievedClassParameters[ck][0]
				}
			}

			d.Set("class_parameters", computedClassParameters)
			return nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				// Log the error
				tflog.Debug(ctx, fmt.Sprintf("Unable to read information from IPv6 subnet: %s (%s)\n", d.Get("name").(string), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to read information from IPv6 subnet: %s\n", d.Get("name").(string)))
		}

		// Reporting a failure
		return diag.Errorf("Unable to find IPv6 subnet: %s", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}
