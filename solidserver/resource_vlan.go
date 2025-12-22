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

func resourcevlan() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcevlanCreate,
		ReadContext:   resourcevlanRead,
		UpdateContext: resourcevlanUpdate,
		DeleteContext: resourcevlanDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcevlanImportState,
		},

		Description: heredoc.Doc(`
			VLANresource allows to create and manage VLAN(s) and VxLAN(s).
		`),

		Schema: map[string]*schema.Schema{
			"vlan_domain": {
				Type:        schema.TypeString,
				Description: "The name of the vlan Domain.",
				Required:    true,
				ForceNew:    true,
			},
			"vlan_range": {
				Type:        schema.TypeString,
				Description: "The name of the vlan Range.",
				//DiffSuppressFunc:      resourcediffsuppressnull,
				Required: false,
				Optional: true,
				ForceNew: true,
				Default:  "",
			},
			"request_id": {
				Type:             schema.TypeInt,
				Description:      "The optionally requested vlan ID.",
				DiffSuppressFunc: resourcediffsuppress,
				Optional:         true,
				ForceNew:         true,
				Default:          0,
			},
			"vlan_id": {
				Type:        schema.TypeInt,
				Description: "The vlan ID.",
				Computed:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vlan to create.",
				Required:    true,
				ForceNew:    false,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the vlan.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to vlan.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourcevlanCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	var vlanIDs []string = nil

	// Determining if a VLAN ID was submitted in or if we should get one from the VLAN Manager
	if d.Get("request_id").(int) > 0 {
		vlanIDs = []string{strconv.Itoa(d.Get("request_id").(int))}
	} else {
		var vlanErr error = nil

		vlanIDs, vlanErr = vlanidfindfree(d.Get("vlan_domain").(string), meta)

		if vlanErr != nil {
			// Reporting a failure
			return diag.FromErr(vlanErr)
		}
	}

	for i := 0; i < len(vlanIDs); i++ {
		// Building parameters
		parameters := url.Values{}
		parameters.Add("add_flag", "new_only")
		parameters.Add("vlmdomain_name", d.Get("vlan_domain").(string))

		if len(d.Get("vlan_range").(string)) > 0 {
			parameters.Add("vlmrange_name", d.Get("vlan_range").(string))
		}

		parameters.Add("vlmvlan_vlan_id", vlanIDs[i])
		parameters.Add("vlmvlan_name", d.Get("name").(string))

		if s.Version < 730 {
			tflog.Info(ctx, fmt.Sprintf("VLAN class parameters are not supported in SOLIDserver Version (%i)\n", s.Version))
		} else {
			parameters.Add("vlmvlan_class_name", d.Get("class").(string))
			parameters.Add("vlmvlan_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())
		}

		// Sending creation request
		resp, body, err := s.Request("post", "rest/vlm_vlan_add", &parameters)

		if err == nil {
			var buf [](map[string]interface{})
			json.Unmarshal([]byte(body), &buf)

			// Checking the answer
			if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
				if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
					tflog.Debug(ctx, fmt.Sprintf("Created vlan (oid): %s\n", oid))

					vnid, _ := strconv.Atoi(vlanIDs[i])
					d.Set("vlan_id", vnid)
					d.SetId(oid)

					return nil
				}
			} else {
				if len(buf) > 0 {
					if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
						tflog.Debug(ctx, fmt.Sprintf("Failed vlan registration for vlan: %s with vnid: %s (%s)\n", d.Get("name").(string), vlanIDs[i], errMsg))
					} else {
						tflog.Debug(ctx, fmt.Sprintf("Failed vlan registration for vlan: %s with vnid: %s\n", d.Get("name").(string), vlanIDs[i]))
					}
				} else {
					tflog.Debug(ctx, fmt.Sprintf("Failed vlan registration for vlan: %s with vnid: %s\n", d.Get("name").(string), vlanIDs[i]))
				}
			}
		} else {
			// Reporting a failure
			return diag.FromErr(err)
		}
	}

	// Reporting a failure
	return diag.Errorf("Unable to create vlan: %s, unable to find a suitable vnid\n", d.Get("name").(string))
}

func resourcevlanUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmvlan_id", d.Id())
	parameters.Add("add_flag", "edit_only")
	parameters.Add("vlmvlan_name", d.Get("name").(string))

	if s.Version < 730 {
		tflog.Info(ctx, fmt.Sprintf("VLAN class parameters are not supported in SOLIDserver Version (%i)\n", s.Version))
	} else {
		parameters.Add("vlmvlan_class_name", d.Get("class").(string))
		parameters.Add("vlmvlan_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())
	}

	// Sending the update request
	resp, body, err := s.Request("put", "rest/vlm_vlan_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Updated vlan (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to update vlan: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to update vlan: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlanDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmvlan_id", d.Id())

	// Sending the deletion request
	resp, body, err := s.Request("delete", "rest/vlm_vlan_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					return diag.Errorf("Unable to delete vlan: %s (%s)", d.Get("name").(string), errMsg)
				}
			}

			return diag.Errorf("Unable to delete vlan: %s", d.Get("name").(string))
		}

		// Log deletion
		tflog.Debug(ctx, fmt.Sprintf("Deleted vlan (oid): %s\n", d.Id()))

		// Unset local ID
		d.SetId("")

		// Reporting a success
		return nil
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlanRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmvlan_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmvlan_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			vnid, _ := strconv.Atoi(buf[0]["vlmvlan_vlan_id"].(string))

			d.Set("name", buf[0]["vlmvlan_name"].(string))
			/* Do not read vlan_domain nor vlan_range as their names may change
			// At least until suitable option is found to ignore the change
			d.Set("vlan_domain", buf[0]["vlmdomain_name"].(string))
			if buf[0]["vlmrange_name"].(string) != "#" {
				d.Set("vlan_range", buf[0]["vlmrange_name"].(string))
			} else {
				d.Set("vlan_range", "")
			}
			*/
			d.Set("vlan_id", vnid)

			if s.Version < 730 {
				tflog.Info(ctx, fmt.Sprintf("VLAN class parameters are not supported in SOLIDserver Version (%i)\n", s.Version))
			} else {
				d.Set("class", buf[0]["vlmvlan_class_name"].(string))

				// Updating local class_parameters
				currentClassParameters := d.Get("class_parameters").(map[string]interface{})
				retrievedClassParameters, _ := url.ParseQuery(buf[0]["vlmvlan_class_parameters"].(string))
				computedClassParameters := map[string]string{}

				for ck := range currentClassParameters {
					if rv, rvExist := retrievedClassParameters[ck]; rvExist {
						computedClassParameters[ck] = rv[0]
					} else {
						computedClassParameters[ck] = ""
					}
				}

				d.Set("class_parameters", computedClassParameters)
			}

			return nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				// Log the error
				tflog.Debug(ctx, fmt.Sprintf("Unable to find vlan: %s (%s)\n", d.Get("name"), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find vlan (oid): %s\n", d.Id()))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("Unable to find vlan: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcevlanImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("vlmvlan_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/vlmvlan_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			vnid, _ := strconv.Atoi(buf[0]["vlmvlan_vlan_id"].(string))

			d.Set("name", buf[0]["vlmvlan_name"].(string))
			d.Set("vlan_domain", buf[0]["vlmdomain_name"].(string))
			d.Set("vlan_range", buf[0]["vlmrange_name"].(string))
			d.Set("vlan_id", vnid)

			if s.Version < 730 {
				tflog.Info(ctx, fmt.Sprintf("VLAN class parameters are not supported in SOLIDserver Version (%i)\n", s.Version))
			} else {
				d.Set("class", buf[0]["vlmvlan_class_name"].(string))

				// Updating local class_parameters
				currentClassParameters := d.Get("class_parameters").(map[string]interface{})
				retrievedClassParameters, _ := url.ParseQuery(buf[0]["vlmvlan_class_parameters"].(string))
				computedClassParameters := map[string]string{}

				for ck := range currentClassParameters {
					if rv, rvExist := retrievedClassParameters[ck]; rvExist {
						computedClassParameters[ck] = rv[0]
					} else {
						computedClassParameters[ck] = ""
					}
				}

				d.Set("class_parameters", computedClassParameters)
			}

			return []*schema.ResourceData{d}, nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				tflog.Debug(ctx, fmt.Sprintf("Unable to import vlan(oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import vlan (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("SOLIDServer - Unable to find and import vlan (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
