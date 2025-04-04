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
	"path/filepath"
	"regexp"
	"strconv"
)

func resourcenominterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcenominterfaceCreate,
		ReadContext:   resourcenominterfaceRead,
		UpdateContext: resourcenominterfaceUpdate,
		DeleteContext: resourcenominterfaceDelete,
		Importer:      &schema.ResourceImporter{
			//StateContext: resourcenominterfaceImportState,
		},

		Description: heredoc.Doc(`
			Network interface resource allows to create and manage network interfaces attached to networtk objects within the SOLIDserver's
			Network Object Manager (NOM) module, a lightweight network assets repository..
		`),

		Schema: map[string]*schema.Schema{
			"path": {
				Type:         schema.TypeString,
				Description:  "The path of the object the network interface belongs to.",
				ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile("^\\/.*$|^.*\\/$"), "Path must not starts nor ends with a '/'."),
				Required:     true,
				ForceNew:     true,
			},
			"port": {
				Type:        schema.TypeString,
				Description: "The name of the port.",
				Required:    true,
				ForceNew:    true,
			},
			"mac": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"), "Unsupported MAC address format."),
				Required:     true,
				ForceNew:     true,
			},
			"interface": {
				Type:        schema.TypeString,
				Description: "The name of the network interface.",
				Computed:    true,
				Default:     nil,
			},
			"vlan_domain": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "",
			},
			"vlan": {
				Type:     schema.TypeInt,
				ForceNew: true,
				Optional: true,
				Default:  0,
			},
			"addresses": {
				Type:        schema.TypeList,
				Description: "The list of IP addresses associated with the network interface.",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: IsIPAddressOrEmptyString,
				},
			},
		},
	}
}

func createIPInterface(ctx context.Context, d *schema.ResourceData, meta interface{}, parameters url.Values, ifaceName string, addr string) {
	s := meta.(*SOLIDserver)

	iparameters := parameters
	iparameters.Add("add_flag", "new_only")
	iparameters.Add("nomiface_name", ifaceName)
	iparameters.Add("nomiface_hostaddr", addr)

	// Sending creation request
	resp, body, err := s.Request("post", "rest/nom_iface_add", &iparameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Created network interface %s:%s\n", d.Get("path").(string)+"/"+ifaceName, addr))
		} else {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					tflog.Error(ctx, fmt.Sprintf("Unable to create network interface address: %s:%s (%s)", d.Get("path").(string)+"/"+ifaceName, addr, errMsg))
				}
			} else {
				tflog.Error(ctx, fmt.Sprintf("Unable to create network interface address: %s:s\n", d.Get("path").(string)+"/"+ifaceName, addr))
			}
		}
	} else {
		tflog.Error(ctx, fmt.Sprintf("Unable to create network interface address: %s:s\n", d.Get("path").(string)+"/"+ifaceName, addr))
	}
}

func deleteIPInterface(ctx context.Context, d *schema.ResourceData, meta interface{}, parameters url.Values, ifaceName string, addr string) {
	s := meta.(*SOLIDserver)

	iparameters := parameters
	iparameters.Add("nomiface_name", ifaceName)
	iparameters.Add("nomiface_hostaddr", addr)

	// Sending creation request
	resp, body, err := s.Request("delete", "rest/nom_iface_delete", &iparameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 204) && len(buf) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Deleted network interface %s:%s\n", d.Get("path").(string)+"/"+ifaceName, addr))
		} else {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					tflog.Error(ctx, fmt.Sprintf("Unable to delete network interface address: %s:%s (%s)", d.Get("path").(string)+"/"+ifaceName, addr, errMsg))
				}
			} else {
				tflog.Error(ctx, fmt.Sprintf("Unable to delete network interface address: %s:s\n", d.Get("path").(string)+"/"+ifaceName, addr))
			}
		}
	} else {
		tflog.Error(ctx, fmt.Sprintf("Unable to delete network interface address: %s:s\n", d.Get("path").(string)+"/"+ifaceName, addr))
	}
}

func resourcenominterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building global parameters
	gparameters := url.Values{}

	ifaceName := d.Get("port").(string)

	if (d.Get("vlan_domain").(string) == "" && d.Get("vlan").(int) != 0) || (d.Get("vlan_domain").(string) != "" && d.Get("vlan").(int) == 0) {
		return diag.Errorf("Unable to create network interface: %s (Invalid VLAN configuration)\n", ifaceName)
	}

	// Build ifname based on VlanID if provided
	if d.Get("vlan_domain").(string) != "" {
		if d.Get("vlan").(int) != 0 {
			ifaceName = d.Get("port").(string) + "." + strconv.Itoa(d.Get("vlan").(int))
			gparameters.Add("nomiface_vlan_domain", d.Get("vlan_domain").(string))
			gparameters.Add("nomiface_vlan_number", strconv.Itoa(d.Get("vlan").(int)))
		}
	}

	gparameters.Add("nomnetobj_name", filepath.Base(d.Get("path").(string)))
	gparameters.Add("nomfolder_path", filepath.Dir(d.Get("path").(string)))
	gparameters.Add("nomiface_port_name", d.Get("port").(string))
	gparameters.Add("nomiface_port_mac", d.Get("mac").(string))

	// Sending port creation request
	parameters := gparameters
	parameters.Add("add_flag", "new_only")
	resp, body, err := s.Request("post", "rest/nom_iface_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Created network interface port %s (oid): %s\n", d.Get("path").(string)+"/"+d.Get("port").(string), oid))

				portOID, portErr := portidbyinterfaceid(oid, meta)
				if portErr != nil {
					return diag.Errorf("Unable to retrieve network interface port ID: %s (%s)", d.Get("path").(string)+"/"+ifaceName, portErr)
				}

				d.SetId(portOID)
				d.Set("interface", ifaceName)

				// Sending interfaces creation requests
				for _, addr := range d.Get("addresses").([]interface{}) {
					createIPInterface(ctx, d, meta, gparameters, ifaceName, addr.(string))
				}
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to create network interface: %s (%s)", d.Get("path").(string)+"/"+ifaceName, errMsg)
			}
		}

		return diag.Errorf("Unable to create network interface: %s\n", d.Get("path").(string)+"/"+ifaceName)
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenominterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Building global parameters
	gparameters := url.Values{}

	ifaceName := d.Get("port").(string)

	if (d.Get("vlan_domain").(string) == "" && d.Get("vlan").(int) != 0) || (d.Get("vlan_domain").(string) != "" && d.Get("vlan").(int) == 0) {
		return diag.Errorf("Unable to create network interface: %s (Invalid VLAN configuration)\n", ifaceName)
	}

	// Build ifname based on VlanID if provided
	if d.Get("vlan_domain").(string) != "" {
		if d.Get("vlan").(int) != 0 {
			ifaceName = d.Get("port").(string) + "." + strconv.Itoa(d.Get("vlan").(int))
			gparameters.Add("nomiface_vlan_domain", d.Get("vlan_domain").(string))
			gparameters.Add("nomiface_vlan_number", strconv.Itoa(d.Get("vlan").(int)))
		}
	}

	gparameters.Add("nomnetobj_name", filepath.Base(d.Get("path").(string)))
	gparameters.Add("nomfolder_path", filepath.Dir(d.Get("path").(string)))
	gparameters.Add("nomiface_port_name", d.Get("port").(string))
	gparameters.Add("nomiface_port_mac", d.Get("mac").(string))

	// Synchronizing IP addresses
	old, new := d.GetChange("addresses")
	oldaddrs, _ := interfaceSliceToStringSlice(old.([]interface{}))
	newaddrs, _ := interfaceSliceToStringSlice(new.([]interface{}))

	for _, addr := range newaddrs {
		j := stringOffsetInSlice(addr, oldaddrs)

		if j < 0 {
			// Create new IP address
			tflog.Debug(ctx, fmt.Sprintf("New IP address: %s\n", addr))
			createIPInterface(ctx, d, meta, gparameters, ifaceName, addr)
			continue
		}

		oldaddrs = removeOffsetInSlice(j, oldaddrs)
	}

	//Remove remaining IP addresses
	for _, addr := range oldaddrs {
		tflog.Debug(ctx, fmt.Sprintf("Removing IP address: %s\n", addr))
		deleteIPInterface(ctx, d, meta, gparameters, ifaceName, addr)
	}

	// Other interface do not need to be updated
	return nil
}

func resourcenominterfaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	lparameters := url.Values{}
	//whereClause := "nomfolder_path='" + filepath.Dir(d.Get("path").(string)) + "' and nomnetobj_name='" + filepath.Base(d.Get("path").(string)) + "' and \
	//nomiface_port_name='" + d.Get("port").(string) + "' and nomiface_vlan_number='" + strconv.Itoa(d.Get("vlan").(int)) + "'"
	whereClause := "nomport_id='" + d.Id() + "'"
	lparameters.Add("WHERE", whereClause)

	// Sending read request
	resp, body, err := s.Request("get", "rest/nom_iface_list", &lparameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			for _, rawIface := range buf {
				//tflog.Debug(ctx, fmt.Sprintf("Deleting network interface(s) %s\n", rawIface["nomiface_port_name"].(string)+"."+rawIface["nomiface_vlan_number"].(string)))
				tflog.Debug(ctx, fmt.Sprintf("Deleting network interface(s) %s\n", d.Get("interface").(string)))

				// Building parameters
				dparameters := url.Values{}
				dparameters.Add("nomiface_id", rawIface["nomiface_id"].(string))

				// Sending network interface deletion request
				//resp, body, err :=
				s.Request("delete", "rest/nom_iface_delete", &dparameters)

				//FIXME - Handle errors
			}

			return nil
		}
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcenominterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	//whereClause := "nomfolder_path='" + filepath.Dir(d.Get("path").(string)) + "' and nomnetobj_name='" + filepath.Base(d.Get("path").(string)) + "' and\
	// nomiface_port_name='" + d.Get("port").(string) + "' and nomiface_vlan_number='" + strconv.Itoa(d.Get("vlan").(int)) + "'"
	whereClause := "nomport_id='" + d.Id() + "'"
	parameters.Add("WHERE", whereClause)

	// Sending read request
	resp, body, err := s.Request("get", "rest/nom_iface_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {

			addresses := make([]interface{}, 0)

			for _, rawIface := range buf {
				tflog.Debug(ctx, fmt.Sprintf("Reading network interface(s) information from %s\n", rawIface["nomiface_port_name"].(string)+"."+rawIface["nomiface_vlan_number"].(string)))

				if rawIface["nomiface_hostaddr"].(string) != "" {
					addresses = append(addresses, rawIface["nomiface_hostaddr"])
				} else {
					if rawIface["nomiface_hostaddr6"].(string) != "" {
						addresses = append(addresses, rawIface["nomiface_hostaddr6"])
					}
				}

				d.Set("addresses", addresses)
			}

			return nil
		}

		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				// Log the error
				tflog.Debug(ctx, fmt.Sprintf("Unable to find network interface: %s (%s)\n", d.Get("path").(string)+"/"+d.Get("port").(string), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find network interface: %s\n", d.Get("path").(string)+"/"+d.Get("port").(string)))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("Unable to find network interface: %s\n", d.Get("path").(string)+"/"+d.Get("port").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

//func resourcenominterfaceImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
// s := meta.(*SOLIDserver)

// // Building parameters
// parameters := url.Values{}
// parameters.Add("nomfolder_id", d.Id())

// // Sending the read request
// resp, body, err := s.Request("get", "rest/nom_folder_info", &parameters)

// if err == nil {
// 	var buf [](map[string]interface{})
// 	json.Unmarshal([]byte(body), &buf)

// 	// Checking the answer
// 	if resp.StatusCode == 200 && len(buf) > 0 {
// 		d.Set("description", buf[0]["nomfolder_description"].(string))

// 		if nomSpace, vnomSpaceExist := buf[0]["nomfolder_site_name"].(string); vnomSpaceExist && nomSpace != "#" {
// 			d.Set("space", buf[0]["nomfolder_site_name"].(string))
// 		}

// 		d.Set("class", buf[0]["nomfolder_class_name"].(string))

// 		// Updating local class_parameters
// 		currentClassParameters := d.Get("class_parameters").(map[string]interface{})
// 		retrievedClassParameters, _ := url.ParseQuery(buf[0]["nomfolder_class_parameters"].(string))
// 		computedClassParameters := map[string]string{}

// 		for ck := range currentClassParameters {
// 			if rv, rvExist := retrievedClassParameters[ck]; rvExist {
// 				computedClassParameters[ck] = rv[0]
// 			} else {
// 				computedClassParameters[ck] = ""
// 			}
// 		}

// 		d.Set("class_parameters", computedClassParameters)

// 		return []*schema.ResourceData{d}, nil
// 	}

// 	if len(buf) > 0 {
// 		if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
// 			tflog.Debug(ctx, fmt.Sprintf("Unable to import folder(oid): %s (%s)\n", d.Id(), errMsg))
// 		}
// 	} else {
// 		tflog.Debug(ctx, fmt.Sprintf("Unable to find and import folder (oid): %s\n", d.Id()))
// 	}

// 	// Reporting a failure
// 	return nil, fmt.Errorf("Unable to find and import folder (oid): %s\n", d.Id())
// }

// // Reporting a failure
// return nil, err
//}
