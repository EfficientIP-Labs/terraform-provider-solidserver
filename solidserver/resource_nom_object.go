package solidserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"net/url"
	"strconv"
)

func resourcenomobject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcenomobjectCreate,
		ReadContext:   resourcenomobjectRead,
		UpdateContext: resourcenomobjectUpdate,
		DeleteContext: resourcenomobjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcenomobjectImportState,
		},

		Description: heredoc.Doc(`
			Network Object resource allows to create and manage network objects within the SOLIDserver's Network Object Manager (NOM) module,
			a lightweight network assets repository.
		`),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the network object.",
				Required:    true,
				ForceNew:    true,
			},
			"folder_path": {
				Type:        schema.TypeString,
				Description: "The path of the parent folder of the network object.",
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "A short description of the network object.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"type": {
				Type:         schema.TypeString,
				Description:  "The type of the network object. It cannot exceed 16 characters.",
				ValidateFunc: validation.StringLenBetween(0, 16),
				Optional:     true,
				ForceNew:     false,
				Default:      "",
			},
			"state": {
				Type:         schema.TypeString,
				Description:  "The state of the network object. It cannot exceed 16 characters.",
				ValidateFunc: validation.StringLenBetween(0, 16),
				Optional:     true,
				ForceNew:     false,
				Default:      "",
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the network object.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to the network object.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"interface": {
				Type:        schema.TypeSet,
				Description: "A network interface of the network object.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mac": {
							Type:     schema.TypeString,
							Required: true,
						},
						"address": {
							Type:         schema.TypeString,
							ValidateFunc: validation.IsIPAddress,
							Optional:     true,
						},
						"vlan_domain": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"vlan": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
					},
				},
			},
		},
	}
}

func expandInterfaces(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]string, error) {
	s := meta.(*SOLIDserver)

	configs := d.Get("interface")
	ifaces := configs.(*schema.Set).List()
	results := []string{}
	for _, rawiface := range ifaces {
		// region, ok := limit["region"].(string)
		// if !ok {
		// 	return nil, fmt.Errorf("expected region to be string, got %T instead", limit["region"])
		// }

		iface, ok := rawiface.(map[string]interface{})

		if !ok {
			return nil, fmt.Errorf("Expected interface to be a map[string]interface{}, got %T instead", rawiface)
		}

		parameters := url.Values{}
		parameters.Add("add_flag", "new_only")
		parameters.Add("nomiface_port_name", iface["name"].(string))
		parameters.Add("nomiface_port_mac", iface["mac"].(string))
		parameters.Add("nomfolder_path", d.Get("folder_path").(string))
		parameters.Add("nomnetobj_name", d.Get("name").(string))

		//FIXME How to manage aliases ?
		parameters.Add("nomiface_hostaddr", iface["address"].(string))

		// Build ifname based on VlanID if provided
		// Generate an error if no Vlan Domain is provided
		if iface["vlan"] != 0 {
			if iface["vlan_domain"] == "" {
				tflog.Error(ctx, fmt.Sprintf("Unable create network interface, missing vlan_domain for interface: %s/%s/%s\n", d.Get("folder_path").(string), d.Get("name").(string), iface["name"].(string) + "." + strconv.Itoa(iface["vlan"].(int))))
				continue
			}

			parameters.Add("nomiface_name", iface["address"].(string)+"."+strconv.Itoa(iface["vlan"].(int)))
			parameters.Add("nomiface_vlan_domain", iface["vlan_domain"].(string))
			parameters.Add("nomiface_vlan_number", strconv.Itoa(iface["vlan"].(int)))
		} else {
			parameters.Add("nomiface_name", iface["address"].(string))
		}

		// Sending creation request
		resp, body, err := s.Request("post", "rest/nom_iface_add", &parameters)

		if err != nil {
			tflog.Error(ctx, fmt.Sprintf("Unable create network object: %s\n", err.Error()))
			continue
		}

		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Created network object interface: %s\n", iface["name"].(string)))
		} else {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					tflog.Error(ctx, fmt.Sprintf("Unable create network object: %s (%s)\n", iface["name"].(string), errMsg))
				}
			}
		}

		results = append(results, iface["name"].(string))
	}
	return results, nil
}

func resourcenomobjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("add_flag", "new_only")
	parameters.Add("nomnetobj_name", d.Get("name").(string))
	parameters.Add("nomfolder_path", d.Get("folder_path").(string))
	parameters.Add("nomnetobj_description", d.Get("description").(string))
	parameters.Add("nomnetobj_type", d.Get("type").(string))
	parameters.Add("nomnetobj_state", d.Get("state").(string))
	parameters.Add("nomnetobj_class_name", d.Get("class").(string))
	parameters.Add("nomnetobj_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	// Sending creation request
	resp, body, err := s.Request("post", "rest/nom_netobj_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Created network object (oid): %s\n", oid))
				d.SetId(oid)

				expandInterfaces(ctx, d, meta)

				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to create network object: %s/%s (%s)", d.Get("folder_path").(string), d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to create network object: %s/%s\n", d.Get("folder_path").(string), d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenomobjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("nomnetobj_id", d.Id())
	parameters.Add("add_flag", "edit_only")
	parameters.Add("nomnetobj_description", d.Get("description").(string))
	parameters.Add("nomnetobj_type", d.Get("type").(string))
	parameters.Add("nomnetobj_state", d.Get("state").(string))
	parameters.Add("nomnetobj_class_name", d.Get("class").(string))
	parameters.Add("nomnetobj_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	// Sending the update request
	resp, body, err := s.Request("put", "rest/nom_netobj_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Updated network object (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to update network object: %s/%s (%s)", d.Get("folder_path").(string), d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to update network object: %s/%s\n", d.Get("folder_path").(string), d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenomobjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("nomnetobj_id", d.Id())

	// Sending the deletion request
	resp, body, err := s.Request("delete", "rest/nom_netobj_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					return diag.Errorf("Unable to delete network object: %s/%s (%s)", d.Get("folder_path").(string), d.Get("name").(string), errMsg)
				}
			}

			return diag.Errorf("Unable to delete network object: %s/%s", d.Get("folder_path").(string), d.Get("name").(string))
		}

		// Log deletion
		tflog.Debug(ctx, fmt.Sprintf("Deleted network object (oid): %s\n", d.Id()))

		// Unset local ID
		d.SetId("")

		// Reporting a success
		return nil
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenomobjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("nomnetobj_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/nom_netobj_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("description", buf[0]["nomnetobj_description"].(string))
			d.Set("type", buf[0]["nomnetobj_type"].(string))
			d.Set("state", buf[0]["nomnetobj_state"].(string))
			d.Set("class", buf[0]["nomnetobj_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["nomnetobj_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to find network object: %s/%s (%s)\n", d.Get("folder_path").(string), d.Get("name"), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find network object (oid): %s\n", d.Id()))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("Unable to find network object: %s/%s\n", d.Get("folder_path").(string), d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenomobjectImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("nomfolder_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/nom_netobj_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("description", buf[0]["nomnetobj_description"].(string))
			d.Set("type", buf[0]["nomnetobj_type"].(string))
			d.Set("state", buf[0]["nomnetobj_state"].(string))
			d.Set("class", buf[0]["nomnetobj_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["nomnetobj_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to import network object(oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import network object (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("Unable to find and import network object (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
