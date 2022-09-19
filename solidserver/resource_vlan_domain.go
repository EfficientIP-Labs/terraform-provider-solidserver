package solidserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/url"
	"strconv"
)

func resourcevlandomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcevlandomainCreate,
		ReadContext:   resourcevlandomainRead,
		UpdateContext: resourcevlandomainUpdate,
		DeleteContext: resourcevlandomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcevlandomainImportState,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the VLAN Domain to create.",
				Required:    true,
				ForceNew:    true,
			},
			"vxlan": {
				Type:        schema.TypeBool,
				Description: "Specify if the VLAN Domain is a VXLAN Domain.",
				Optional:    true,
				ForceNew:    true,
				Default:     false,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the VLAN Domain.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to VLAN Domain.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourcevlandomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("add_flag", "new_only")
	parameters.Add("vlmdomain_name", d.Get("name").(string))
	parameters.Add("vlmdomain_class_name", d.Get("class").(string))
	parameters.Add("vlmdomain_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	if d.Get("vxlan").(bool) {
		if s.Version < 700 {
			return diag.Errorf("VXLAN Domain are not supported in this SOLIDserver version %d\n", s.Version)
		}

		parameters.Add("support_vxlan", "1")
	}

	// Sending creation request
	resp, body, err := s.Request("post", "rest/vlm_domain_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Created VLAN Domain (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to create VLAN Domain: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to create VLAN Domain: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlandomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmdomain_id", d.Id())
	parameters.Add("add_flag", "edit_only")
	parameters.Add("vlmdomain_name", d.Get("name").(string))
	parameters.Add("vlmdomain_class_name", d.Get("class").(string))
	parameters.Add("vlmdomain_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	if d.Get("vxlan").(bool) {
		if s.Version < 700 {
			return diag.Errorf("VXLAN Domain are not supported in this SOLIDserver version %d\n", s.Version)
		}
		parameters.Add("support_vxlan", "1")
	}

	// Sending the update request
	resp, body, err := s.Request("put", "rest/vlm_domain_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Updated VLAN Domain (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to VLAN Domain: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to VLAN Domain: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlandomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmdomain_id", d.Id())

	// Sending the deletion request
	resp, body, err := s.Request("delete", "rest/vlm_domain_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					return diag.Errorf("Unable to delete VLAN Domain: %s (%s)", d.Get("name").(string), errMsg)
				}
			}

			return diag.Errorf("Unable to delete VLAN Domain: %s", d.Get("name").(string))
		}

		// Log deletion
		tflog.Debug(ctx, fmt.Sprintf("Deleted VLAN Domain (oid): %s\n", d.Id()))

		// Unset local ID
		d.SetId("")

		// Reporting a success
		return nil
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlandomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmdomain_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmdomain_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			vxlanSupport := false

			if _, exist := buf[0]["support_vxlan"]; exist {
				vxlanSupport, _ = strconv.ParseBool(buf[0]["support_vxlan"].(string))
			}

			d.Set("name", buf[0]["vlmdomain_name"].(string))
			d.Set("support_vxlan", vxlanSupport)
			d.Set("class", buf[0]["vlmdomain_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["vlmdomain_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to find VLAN Domain: %s (%s)\n", d.Get("name"), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find VLAN Domain (oid): %s\n", d.Id()))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("Unable to find VLAN Domain: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlandomainImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmdomain_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmdomain_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			vxlanSupport := false

			if _, exist := buf[0]["support_vxlan"]; exist {
				vxlanSupport, _ = strconv.ParseBool(buf[0]["support_vxlan"].(string))
			}

			d.Set("name", buf[0]["vlmdomain_name"].(string))
			d.Set("support_vxlan", vxlanSupport)
			d.Set("class", buf[0]["vlmdomain_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["vlmdomain_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to import VLAN Domain(oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import VLAN Domain (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("SOLIDServer - Unable to find and import VLAN Domain (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
