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

func resourcevlanrange() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcevlanrangeCreate,
		ReadContext:   resourcevlanrangeRead,
		UpdateContext: resourcevlanrangeUpdate,
		DeleteContext: resourcevlanrangeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcevlanrangeImportState,
		},

		Description: heredoc.Doc(`
			VLAN Range resource allows to create and manage VLAN and VxLAN ranges.
		`),

		Schema: map[string]*schema.Schema{
			"vlan_domain": {
				Type:        schema.TypeString,
				Description: "The name of the vlan Domain.",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the VLAN Range to create.",
				Required:    true,
				ForceNew:    true,
			},
			"start": {
				Type:        schema.TypeInt,
				Description: "The vlan range's lower vlan ID.",
				Required:    true,
				ForceNew:    true,
			},
			"end": {
				Type:        schema.TypeInt,
				Description: "The vlan range's higher vlan ID.",
				Required:    true,
				ForceNew:    true,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the VLAN Range.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to VLAN Range.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourcevlanrangeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("add_flag", "new_only")
	parameters.Add("vlmdomain_name", d.Get("vlan_domain").(string))
	parameters.Add("vlmrange_name", d.Get("name").(string))
	parameters.Add("vlmrange_start_vlan_id", strconv.Itoa(d.Get("start").(int)))
	parameters.Add("vlmrange_end_vlan_id", strconv.Itoa(d.Get("end").(int)))
	parameters.Add("vlmrange_class_name", d.Get("class").(string))
	parameters.Add("vlmrange_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	// Sending creation request
	resp, body, err := s.Request("post", "rest/vlm_range_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Created VLAN Range (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to create VLAN Range: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to create VLAN Range: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlanrangeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmrange_id", d.Id())
	parameters.Add("add_flag", "edit_only")
	parameters.Add("vlmrange_name", d.Get("name").(string))
	parameters.Add("vlmrange_class_name", d.Get("class").(string))
	parameters.Add("vlmrange_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	// Sending the update request
	resp, body, err := s.Request("put", "rest/vlm_range_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Updated VLAN Range (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to VLAN Range: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to VLAN Range: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlanrangeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmrange_id", d.Id())

	// Sending the deletion request
	resp, body, err := s.Request("delete", "rest/vlm_range_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					return diag.Errorf("Unable to delete VLAN Range: %s (%s)", d.Get("name").(string), errMsg)
				}
			}

			return diag.Errorf("Unable to delete VLAN Range: %s", d.Get("name").(string))
		}

		// Log deletion
		tflog.Debug(ctx, fmt.Sprintf("Deleted VLAN Range (oid): %s\n", d.Id()))

		// Unset local ID
		d.SetId("")

		// Reporting a success
		return nil
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlanrangeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmrange_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmrange_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			vxlanSupport := false

			if _, exist := buf[0]["support_vxlan"]; exist {
				vxlanSupport, _ = strconv.ParseBool(buf[0]["support_vxlan"].(string))
			}

			d.Set("name", buf[0]["vlmrange_name"].(string))
			d.Set("support_vxlan", vxlanSupport)
			d.Set("class", buf[0]["vlmrange_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["vlmrange_class_parameters"].(string))
			computedClassParameters := map[string]string{}

			for ck := range currentClassParameters {
				if rv, rvExist := retrievedClassParameters[ck]; rvExist {
					computedClassParameters[ck] = rv[0]
				} else {
					computedClassParameters[ck] = ""
				}
			}

			d.Set("class_parameters", computedClassParameters)

			return nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				// Log the error
				tflog.Debug(ctx, fmt.Sprintf("Unable to find VLAN Range: %s (%s)\n", d.Get("name"), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find VLAN Range (oid): %s\n", d.Id()))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("Unable to find VLAN Range: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlanrangeImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmrange_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmrange_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			vxlanSupport := false

			if _, exist := buf[0]["support_vxlan"]; exist {
				vxlanSupport, _ = strconv.ParseBool(buf[0]["support_vxlan"].(string))
			}

			d.Set("name", buf[0]["vlmrange_name"].(string))
			d.Set("support_vxlan", vxlanSupport)
			d.Set("class", buf[0]["vlmrange_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["vlmrange_class_parameters"].(string))
			computedClassParameters := map[string]string{}

			for ck := range currentClassParameters {
				if rv, rvExist := retrievedClassParameters[ck]; rvExist {
					computedClassParameters[ck] = rv[0]
				} else {
					computedClassParameters[ck] = ""
				}
			}

			d.Set("class_parameters", computedClassParameters)

			return []*schema.ResourceData{d}, nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(ctx, fmt.Sprintf("Unable to import VLAN Range(oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import VLAN Range (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("SOLIDServer - Unable to find and import VLAN Range (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
