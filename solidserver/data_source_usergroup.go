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

func dataSourceusergroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceusergroupRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the user group.",
				Required:    true,
			},
		},
	}
}

func dataSourceusergroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)
	d.SetId("")

	// Building parameters
	parameters := url.Values{}
	parameters.Add("WHERE", "grp_name='"+d.Get("name").(string)+"'")

	// Sending the read request
	resp, body, err := s.Request("get", "rest/group_admin_list", &parameters)

	if err != nil {
		return diag.Errorf("Error on group %s %s\n", d.Get("name").(string), err)
	}

	var buf [](map[string]interface{})
	json.Unmarshal([]byte(body), &buf)

	// Checking the answer
	if resp.StatusCode == 200 && len(buf) > 0 {
		d.SetId(buf[0]["grp_id"].(string))

		return nil
	}

	if len(buf) > 0 {
		if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find group: %s (%s)\n", d.Get("name").(string), errMsg))
		}
	} else {
		// Log the error
		return diag.Errorf("Unable to find group: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.Errorf("Error retreiving group : %s\n", d.Get("name").(string))
}
