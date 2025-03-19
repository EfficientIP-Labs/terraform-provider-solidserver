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
	"regexp"
	"strings"
)

func resourceipaddress() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceipaddressCreate,
		ReadContext:   resourceipaddressRead,
		UpdateContext: resourceipaddressUpdate,
		DeleteContext: resourceipaddressDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceipaddressImportState,
		},

		Description: heredoc.Doc(`
			IP address resource allows to create and manage reserved addresses for specific devices, apps or users.
			More importantly it allows to store useful meta-data for both tracking and automation purposes.
		`),

		Schema: map[string]*schema.Schema{
			"space": {
				Type:        schema.TypeString,
				Description: "The name of the space into which creating the IP address.",
				Required:    true,
				ForceNew:    true,
			},
			"subnet": {
				Type:        schema.TypeString,
				Description: "The name of the subnet into which creating the IP address.",
				Required:    true,
				ForceNew:    true,
			},
			"pool": {
				Type:        schema.TypeString,
				Description: "The name of the pool into which creating the IP address.",
				Optional:    true,
				ForceNew:    true,
				Default:     "",
			},
			"request_ip": {
				Type:         schema.TypeString,
				Description:  "The optionally requested IP address.",
				ValidateFunc: validation.IsIPAddress,
				Optional:     true,
				ForceNew:     true,
				Default:      "",
			},
			"assignment_order": {
				Type:         schema.TypeString,
				Description:  "An optional IP assignment order within the parent subnet/pool (Supported: optimized, start, end; Default: optimized).",
				ValidateFunc: validation.StringInSlice([]string{"optimized", "start", "end"}, false),
				Optional:     true,
				ForceNew:     true,
				Default:      "optimized",
			},
			"address": {
				Type:        schema.TypeString,
				Description: "The provisionned IP address.",
				Computed:    true,
				ForceNew:    true,
			},
			"device": {
				Type:        schema.TypeString,
				Description: "Device Name to associate with the IP address (Require a 'Device Manager' license).",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The short name or FQDN of the IP address to create.",
				Required:    true,
				ForceNew:    false,
			},
			"mac": {
				Type:             schema.TypeString,
				Description:      "The MAC Address of the IP address to create.",
				ValidateFunc:     validation.StringMatch(regexp.MustCompile("^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$"), "Unsupported MAC address format."),
				Optional:         true,
				ForceNew:         false,
				DiffSuppressFunc: resourcediffsuppresscase,
				Default:          "",
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the IP address.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to the IP address.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceipaddressCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	var requestedHexIP string = iptohexip(d.Get("request_ip").(string))
	var poolInfo map[string]interface{} = nil
	var ipAddresses []string = nil
	var deviceID string = ""

	// Gather required ID(s) from provided information
	siteID, siteErr := ipsiteidbyname(d.Get("space").(string), meta)

	if siteErr != nil {
		// Reporting a failure
		return diag.FromErr(siteErr)
	}

	//subnetID, subnetErr := ipsubnetidbyname(siteID, d.Get("subnet").(string), true, meta)
	//if subnetErr != nil {
	//	// Reporting a failure
	//	return diag.FromErr(subnetErr)
	//}

	subnetInfo, subnetErr := ipsubnetinfobyname(siteID, d.Get("subnet").(string), true, meta)

	if subnetInfo == nil || subnetErr != nil {
		// Reporting a failure
		if subnetInfo == nil {
			return diag.Errorf("Unable to create IP address: %s, unable to find requested network\n", d.Get("name").(string))
		}

		return diag.FromErr(subnetErr)
	}

	if len(d.Get("pool").(string)) > 0 {
		var poolErr error = nil

		poolInfo, poolErr = ippoolinfobyname(siteID, d.Get("pool").(string), d.Get("subnet").(string), meta)
		if poolErr != nil {
			// Reporting a failure
			return diag.FromErr(poolErr)
		}
	}

	// Retrieving device ID
	if len(d.Get("device").(string)) > 0 {
		var deviceErr error = nil

		deviceID, deviceErr = hostdevidbyname(d.Get("device").(string), meta)
		if deviceErr != nil {
			// Reporting a failure
			return diag.FromErr(deviceErr)
		}
	}

	// Determining if an IP address was submitted in or if we should get one from the IPAM
	if len(d.Get("request_ip").(string)) > 0 {
		// Ensure IP Address is within the given subnet start and end IP addresses
		if strings.Compare(subnetInfo["terminal"].(string), "1") == 0 &&
			strings.Compare(subnetInfo["start_hex_addr"].(string), requestedHexIP) == -1 &&
			strings.Compare(requestedHexIP, subnetInfo["end_hex_addr"].(string)) == -1 {

			if poolInfo != nil && (strings.Compare(poolInfo["start_hex_addr"].(string), requestedHexIP) == 1 ||
				strings.Compare(requestedHexIP, poolInfo["end_hex_addr"].(string)) == 1) {
				return diag.Errorf("Unable to create IP address: %s, address is out of pool's range\n", d.Get("name").(string))
			}

			ipAddresses = []string{d.Get("request_ip").(string)}
		} else {
			return diag.Errorf("Unable to create IP address: %s, address is out of network's range\n", d.Get("name").(string))
		}
	} else {
		var poolID string = ""
		var ipErr error = nil

		if poolInfo != nil {
			poolID = poolInfo["id"].(string)
		}

		ipAddresses, ipErr = ipaddressfindfree(subnetInfo["id"].(string), poolID, d.Get("assignment_order").(string), meta)

		if ipErr != nil {
			// Reporting a failure
			return diag.FromErr(ipErr)
		}
	}

	for i := 0; i < len(ipAddresses); i++ {
		// Building parameters
		parameters := url.Values{}
		parameters.Add("site_id", siteID)
		parameters.Add("add_flag", "new_only")
		parameters.Add("ip_name", d.Get("name").(string))
		parameters.Add("hostaddr", ipAddresses[i])
		parameters.Add("hostdev_id", deviceID)
		parameters.Add("ip_class_name", d.Get("class").(string))

		if d.Get("mac").(string) != "" {
			parameters.Add("mac_addr", d.Get("mac").(string))
		}

		// Building class_parameters
		parameters.Add("ip_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

		// Sending the creation request
		resp, body, err := s.Request("post", "rest/ip_add", &parameters)

		if err == nil {
			var buf [](map[string]interface{})
			json.Unmarshal([]byte(body), &buf)

			// Checking the answer
			if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
				if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
					tflog.Debug(ctx, fmt.Sprintf("Created IP address (oid): %s\n", oid))
					d.SetId(oid)
					d.Set("address", ipAddresses[i])
					return nil
				}
			} else {
				if len(buf) > 0 {
					if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
						tflog.Debug(ctx, fmt.Sprintf("Failed IP address registration for IP address: %s with address: %s (%s)\n", d.Get("name").(string), ipAddresses[i], errMsg))
					} else {
						tflog.Debug(ctx, fmt.Sprintf("Failed IP address registration for IP address: %s with address: %s\n", d.Get("name").(string), ipAddresses[i]))
					}
				} else {
					tflog.Debug(ctx, fmt.Sprintf("Failed IP address registration for IP address: %s with address: %s\n", d.Get("name").(string), ipAddresses[i]))
				}
			}
		} else {
			// Reporting a failure
			return diag.FromErr(err)
		}
	}

	// Reporting a failure
	return diag.Errorf("Unable to create IP address: %s, unable to find a suitable network or address\n", d.Get("name").(string))
}

func resourceipaddressUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	var deviceID string = ""

	// Retrieving device ID
	if len(d.Get("device").(string)) > 0 {
		var err error = nil

		deviceID, err = hostdevidbyname(d.Get("device").(string), meta)

		if err != nil {
			// Reporting a failure
			return diag.FromErr(err)
		}
	}

	// Building parameters
	parameters := url.Values{}
	parameters.Add("ip_id", d.Id())
	parameters.Add("add_flag", "edit_only")
	parameters.Add("ip_name", d.Get("name").(string))
	parameters.Add("hostdev_id", deviceID)
	parameters.Add("ip_class_name", d.Get("class").(string))

	if d.Get("mac").(string) != "" {
		parameters.Add("mac_addr", d.Get("mac").(string))
	}

	// Building class_parameters
	parameters.Add("ip_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	// Sending the update request
	resp, body, err := s.Request("put", "rest/ip_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Updated IP address (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to update IP address: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to update IP address: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourceipaddressDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("ip_id", d.Id())

	// Sending the deletion request
	resp, body, err := s.Request("delete", "rest/ip_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					return diag.Errorf("Unable to delete IP address : %s (%s)", d.Get("name").(string), errMsg)
				}
			}

			return diag.Errorf("Unable to delete IP address : %s", d.Get("name").(string))
		}

		// Log deletion
		tflog.Debug(ctx, fmt.Sprintf("Deleted IP address's oid: %s\n", d.Id()))

		// Unset local ID
		d.SetId("")

		// Reporting a success
		return nil
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourceipaddressRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("ip_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_address_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("space", buf[0]["site_name"].(string))
			d.Set("subnet", buf[0]["subnet_name"].(string))
			d.Set("address", hexiptoip(buf[0]["ip_addr"].(string)))
			d.Set("name", buf[0]["name"].(string))

			if macIgnore, _ := regexp.MatchString("^EIP:", buf[0]["mac_addr"].(string)); !macIgnore {
				d.Set("mac", buf[0]["mac_addr"].(string))
			} else {
				d.Set("mac", "")
			}

			d.Set("class", buf[0]["ip_class_name"].(string))
			d.Set("pool", buf[0]["pool_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["ip_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to find IP address: %s (%s)\n", d.Get("name"), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find IP address (oid): %s\n", d.Id()))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("Unable to find IP address: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourceipaddressImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("ip_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/ip_address_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("space", buf[0]["site_name"].(string))
			d.Set("subnet", buf[0]["subnet_name"].(string))
			d.Set("address", hexiptoip(buf[0]["ip_addr"].(string)))
			d.Set("name", buf[0]["name"].(string))
			d.Set("mac", buf[0]["mac_addr"].(string))
			d.Set("class", buf[0]["ip_class_name"].(string))
			d.Set("pool", buf[0]["pool_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["ip_class_parameters"].(string))
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
				// Log the error
				tflog.Debug(ctx, fmt.Sprintf("Unable to import IP address (oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import IP address (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("SOLIDServer - Unable to find and import IP address (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
