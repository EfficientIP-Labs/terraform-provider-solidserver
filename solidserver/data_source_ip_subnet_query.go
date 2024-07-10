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

func dataSourceipsubnetquery() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceipsubnetqueryRead,

		Description: heredoc.Doc(`
			IP subnet query data-source allows to retrieve information about the first IPv4 subnet matching given criterias, including its meta-data.
		`),

		Schema: map[string]*schema.Schema{
			"query": {
				Type:        schema.TypeString,
				Description: "The query used to find the first matching subnet.",
				Required:    true,
			},
			"tags": {
				Type:        schema.TypeString,
				Description: "The tags to be used to find the first matching subnet in the query.",
				Optional:    true,
				Default:     "",
			},
			"orderby": {
				Type:        schema.TypeString,
				Description: "The query used to find the first matching subnet.",
				Optional:    true,
				Default:     "",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the IP subnet.",
				Computed:    true,
			},
			"space": {
				Type:        schema.TypeString,
				Description: "The space associated to the IP subnet.",
				Computed:    true,
			},
			"address": {
				Type:        schema.TypeString,
				Description: "The IP subnet address.",
				Computed:    true,
			},
			"prefix": {
				Type:        schema.TypeString,
				Description: "The IP subnet prefix.",
				Computed:    true,
			},
			"prefix_size": {
				Type:        schema.TypeInt,
				Description: "The IP subnet's prefix length (ex: 24 for a '/24').",
				Computed:    true,
			},
			"netmask": {
				Type:        schema.TypeString,
				Description: "The IP subnet netmask.",
				Computed:    true,
			},
			"terminal": {
				Type:        schema.TypeBool,
				Description: "The terminal property of the IP subnet.",
				Computed:    true,
			},
			"vlan_domain": {
				Type:        schema.TypeString,
				Description: "The optional vlan Domain associated with the subnet.",
				Computed:    true,
			},
			"vlan_range": {
				Type:        schema.TypeString,
				Description: "The optional vlan Range associated with the subnet.",
				Computed:    true,
			},
			"vlan_id": {
				Type:        schema.TypeInt,
				Description: "The optional vlan ID associated with the subnet.",
				Computed:    true,
			},
			"vlan_name": {
				Type:        schema.TypeString,
				Description: "The optional vlan Name associated with the subnet.",
				Computed:    true,
			},
			"gateway": {
				Type:        schema.TypeString,
				Description: "The subnet's computed gateway.",
				Computed:    true,
				ForceNew:    true,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the IP subnet.",
				Computed:    true,
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to IP subnet.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceipsubnetqueryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)
	d.SetId("")

	// Building parameters
	parameters := url.Values{}
	parameters.Add("TAGS", d.Get("tags").(string))
	parameters.Add("WHERE", d.Get("query").(string))
	parameters.Add("ORDERBY", d.Get("orderby").(string))
	parameters.Add("limit", "1")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_block_subnet_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.SetId(buf[0]["subnet_id"].(string))

			address := hexiptoip(buf[0]["start_ip_addr"].(string))
			subnet_size, _ := strconv.Atoi(buf[0]["subnet_size"].(string))
			prefix_length := sizetoprefixlength(subnet_size)
			prefix := address + "/" + strconv.Itoa(prefix_length)

			d.Set("name", buf[0]["subnet_name"].(string))
			d.Set("address", address)
			d.Set("prefix", prefix)
			d.Set("prefix_size", prefix_length)
			d.Set("netmask", prefixlengthtohexip(prefix_length))

			if buf[0]["is_terminal"].(string) == "1" {
				d.Set("terminal", true)
			} else {
				d.Set("terminal", false)
			}

			if vlanDomain, vlanDomainExist := buf[0]["vlmdomain_name"].(string); vlanDomainExist && vlanDomain != "#" {
				d.Set("vlan_domain", vlanDomain)
			}

			if vlanRange, vlanRangeExist := buf[0]["vlmrange_name"].(string); vlanRangeExist && vlanRange != "#" {
				d.Set("vlan_range", vlanRange)
			}

			if vlanID, vlanIDExist := buf[0]["vlmvlan_vlan_id"].(string); vlanIDExist && vlanID != "0" {
				vlanID, _ := strconv.Atoi(vlanID)
				d.Set("vlan_id", vlanID)
			}

			if vlanName, vlanNameExist := buf[0]["vlmvlan_name"].(string); vlanNameExist && vlanName != "" {
				d.Set("vlan_name", vlanName)
			}

			d.Set("class", buf[0]["subnet_class_name"].(string))

			// Setting local class_parameters
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["subnet_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to read information from IP subnet: %s (%s)\n", d.Get("name").(string), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to read information from IP subnet: %s\n", d.Get("name").(string)))
		}

		// Reporting a failure
		return diag.Errorf("Unable to find IP subnet: %s", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}
