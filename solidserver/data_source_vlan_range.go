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

func dataSourcevlanrange() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcevlanrangeRead,

		Description: heredoc.Doc(`
			VLAN range data-source allows to retrieve information about VLAN ranges, including meta-data.
		`),

		Schema: map[string]*schema.Schema{
			"vlan_domain": {
				Type:        schema.TypeString,
				Description: "The name of the vlan Domain.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the VLAN Range.",
				Required:    true,
			},
			"start": {
				Type:        schema.TypeInt,
				Description: "The vlan range's lower vlan ID.",
				Computed:    true,
			},
			"end": {
				Type:        schema.TypeInt,
				Description: "The vlan range's higher vlan ID.",
				Computed:    true,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the VLAN Range.",
				Computed:    true,
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to VLAN Range.",
				Computed:    true,
			},
		},
	}
}

func dataSourcevlanrangeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)
	d.SetId("")

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "vlmdomain_name='"+d.Get("vlan_domain").(string)+"' AND vlmrange_name='"+d.Get("name").(string)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmrange_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.SetId(buf[0]["vlmrange_id"].(string))

			d.Set("name", buf[0]["vlmrange_name"].(string))

			start, _ := strconv.Atoi(buf[0]["vlmrange_start_vlan_id"].(string))
			end, _ := strconv.Atoi(buf[0]["vlmrange_end_vlan_id"].(string))

			d.Set("start", start)
			d.Set("end", end)

			d.Set("class", buf[0]["vlmrange_class_name"].(string))

			// Updating local class_parameters
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["vlmrange_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to read information from VLAN Range: %s (%s)\n", d.Get("name").(string), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to read information from VLAN Range: %s\n", d.Get("name").(string)))
		}

		// Reporting a failure
		return diag.Errorf("Unable to find VLAN Range: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}
