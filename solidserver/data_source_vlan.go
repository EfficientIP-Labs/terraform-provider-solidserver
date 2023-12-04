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

func dataSourcevlan() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcevlanRead,

		Description: heredoc.Doc(`
			VLAN data-source allows to retrieve information about vlans, including meta-data.
		`),

		Schema: map[string]*schema.Schema{
			"vlan_domain": {
				Type:        schema.TypeString,
				Description: "The name of the vlan Domain.",
				Required:    true,
			},
			"vlan_range": {
				Type:        schema.TypeString,
				Description: "The name of the vlan Range.",
				Required:    false,
				Optional:    true,
				Default:     "",
			},
			"vlan_id": {
				Type:        schema.TypeInt,
				Description: "The vlan ID.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vlan.",
				Required:    true,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the vlan.",
				Computed:    true,
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to vlan.",
				Computed:    true,
			},
		},
	}
}

func dataSourcevlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)
	d.SetId("")

	// Building parameters
	parameters := url.Values{}

	whereClause := "vlmdomain_name='" + d.Get("vlan_domain").(string) + "' AND vlmvlan_name='" + d.Get("name").(string) + "'"

	if vlanRange, ok := d.Get("vlan_range").(string); ok && vlanRange != "" {
		whereClause += " AND vlmrange_name='" + vlanRange + "'"
	}

	parameters.Add("WHERE", whereClause)

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmvlan_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.SetId(buf[0]["vlmvlan_id"].(string))

			d.Set("vlan_range", buf[0]["vlmrange_name"].(string))
			d.Set("name", buf[0]["vlmvlan_name"].(string))

			vlanID, _ := strconv.Atoi(buf[0]["vlmvlan_vlan_id"].(string))
			d.Set("vlan_id", vlanID)

			d.Set("class", buf[0]["vlmvlan_class_name"].(string))

			// Updating local class_parameters
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["vlmvlan_class_parameters"].(string))
			computedClassParameters := map[string]string{}

			for ck := range retrievedClassParameters {
				computedClassParameters[ck] = retrievedClassParameters[ck][0]
			}

			d.Set("class_parameters", computedClassParameters)
			return nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				// Log the error
				tflog.Debug(ctx, fmt.Sprintf("Unable to read information from VLAN: %s (%s)\n", d.Get("name").(string), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to read information from VLAN: %s\n", d.Get("name").(string)))
		}

		// Reporting a failure
		return diag.Errorf("Unable to find VLAN: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}
