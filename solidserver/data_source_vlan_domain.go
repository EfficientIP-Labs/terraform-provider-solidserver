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

func dataSourcevlandomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcevlandomainRead,

		Description: heredoc.Doc(`
			VLAN domain data-source allows to retrieve information about VLAN Domains, including meta-data.
		`),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the VLAN Domain.",
				Required:    true,
				ForceNew:    true,
			},
			"vxlan": {
				Type:        schema.TypeBool,
				Description: "Specify if the VLAN Domain is a VXLAN Domain.",
				Computed:    true,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the VLAN Domain.",
				Computed:    true,
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to VLAN Domain.",
				Computed:    true,
			},
		},
	}
}

func dataSourcevlandomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)
	d.SetId("")

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "vlmdomain_name='"+d.Get("name").(string)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmdomain_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.SetId(buf[0]["vlmdomain_id"].(string))

			d.Set("name", buf[0]["vlmdomain_name"].(string))

			vxlanSupport := false

			if _, exist := buf[0]["support_vxlan"]; exist {
				vxlanSupport, _ = strconv.ParseBool(buf[0]["support_vxlan"].(string))
			}

			d.Set("support_vxlan", vxlanSupport)
			d.Set("class", buf[0]["vlmdomain_class_name"].(string))

			// Updating local class_parameters
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["vlmdomain_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to read information from VLAN Domain: %s (%s)\n", d.Get("name").(string), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to read information from VLAN Domain: %s\n", d.Get("name").(string)))
		}

		// Reporting a failure
		return diag.Errorf("Unable to find VLAN Domain: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}
