package solidserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/url"
)

func dataSourceipspace() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceipspaceRead,

		Description: heredoc.Doc(`
			The space is the highest level in the SOLIDserver's IPAM module organization,
			the entry point of any IPv4 or IPv6 addressing plan.
			It allows to manage unique ranges of IP addresses.
		`),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the IP space.",
				Required:    true,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the IP space.",
				Computed:    true,
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to IP space.",
				Computed:    true,
			},
		},
	}
}

func dataSourceipspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)
	d.SetId("")

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "site_name='"+d.Get("name").(string)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_site_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.SetId(buf[0]["site_id"].(string))

			d.Set("name", buf[0]["site_name"].(string))
			d.Set("class", buf[0]["site_class_name"].(string))

			// Updating local class_parameters
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["site_class_parameters"].(string))
			computedClassParameters := map[string]string{}

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
				tflog.Debug(ctx, fmt.Sprintf("Unable to read information from IP space: %s (%s)\n", d.Get("name").(string), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to read information from IP space: %s\n", d.Get("name").(string)))
		}

		// Reporting a failure
		return diag.Errorf("Unable to find IP space: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}
