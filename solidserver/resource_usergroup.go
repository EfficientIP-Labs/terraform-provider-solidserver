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
	// "strconv"
)

func resourceusergroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceusergroupCreate,
		ReadContext:   resourceusergroupRead,
		UpdateContext: resourceusergroupUpdate,
		DeleteContext: resourceusergroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceusergroupImportState,
		},

		Description: heredoc.Doc(`
			User resource allows to associate users with groups managing
			SOLIDserver permissions.
		`),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the group",
				Required:    true,
				ForceNew:    false,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the group",
				Required:    false,
				Optional:    true,
				ForceNew:    false,
			},
		},
	}
}

func resourceusergroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("add_flag", "new_only")
	parameters.Add("grp_name", d.Get("name").(string))

	if len(d.Get("description").(string)) > 0 {
		parameters.Add("grp_description", d.Get("description").(string))
	}

	// Sending creation request of the user
	resp, body, err := s.Request("post", "rest/group_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Created group (oid): %s\n", oid))
				d.SetId(oid)
			}
		}
	} else {
		return diag.FromErr(err)
	}

	return nil
}

func resourceusergroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("grp_id", d.Id())
	parameters.Add("add_flag", "edit_only")

	bChange := false

	// check for modification on the user
	aVars := map[string]string{
		"description": "grp_description",
		"name":        "grp_name",
	}

	for k, v := range aVars {
		a, b := d.GetChange(k)
		if a != b {
			bChange = true
			parameters.Add(v, b.(string))
		}
	}

	if bChange {
		// Sending the update request
		resp, body, err := s.Request("put", "rest/group_add", &parameters)

		if err == nil {
			var buf [](map[string]interface{})
			json.Unmarshal([]byte(body), &buf)

			// Checking the answer
			if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
				if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
					tflog.Debug(ctx, fmt.Sprintf("Updated group (oid): %s\n", oid))
					d.SetId(oid)
				}
			} else {
				return diag.Errorf("Unable to update group: %s\n", d.Get("name").(string))
			}
		} else {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceusergroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("grp_id", d.Id())

	// Sending the deletion request
	resp, body, err := s.Request("delete", "rest/group_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Group deleted (oid): %s\n", oid))
				d.SetId("")
				return nil
			}
		}
	}

	return diag.Errorf("error deleting group (oid): %s\n", d.Id())
}

func resourceusergroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("grp_id", d.Id())

	// Sending read request
	resp, body, err := s.Request("get", "rest/group_admin_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking answer
		if (resp.StatusCode == 200) && len(buf) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Found group (oid): %s\n", d.Id()))

			d.Set("description", buf[0]["grp_description"].(string))
			d.Set("name", buf[0]["grp_name"].(string))

			return nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to find group %s: %s\n",
					d.Id(),
					errMsg)
			}
		}
	}

	return diag.Errorf("Unable to find group (oid): %s\n", d.Id())
}

func resourceusergroupImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("grp_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/group_admin_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("name", buf[0]["grp_name"].(string))
			d.Set("description", buf[0]["grp_description"].(string))

			return []*schema.ResourceData{d}, nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(ctx, fmt.Sprintf("Unable to import group (oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import group (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("Unable to find and import group (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
