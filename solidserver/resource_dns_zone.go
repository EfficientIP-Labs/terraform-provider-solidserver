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

func resourcednszone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcednszoneCreate,
		ReadContext:   resourcednszoneRead,
		UpdateContext: resourcednszoneUpdate,
		DeleteContext: resourcednszoneDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcednszoneImportState,
		},

		Description: heredoc.Doc(`
			DNS Zone resource allows to create and configure DNS zones.
		`),

		Schema: map[string]*schema.Schema{
			"dnsserver": {
				Type:        schema.TypeString,
				Description: "The name of DNS server or DNS SMART hosting the DNS zone to create.",
				Required:    true,
				ForceNew:    true,
			},
			"dnsview": {
				Type:        schema.TypeString,
				Description: "The name of DNS view hosting the DNS zone to create.",
				Optional:    true,
				ForceNew:    true,
				Default:     "#",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The Domain Name to be hosted by the zone.",
				Required:    true,
				ForceNew:    true,
			},
			"space": {
				Type:        schema.TypeString,
				Description: "The name of the IP space associated to the zone.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"type": {
				Type:         schema.TypeString,
				Description:  "The type of the zone to create (Supported: Master).",
				ValidateFunc: resourcednszonevalidatetype,
				Optional:     true,
				ForceNew:     true,
				Default:      "Master",
			},
			"createptr": {
				Type:        schema.TypeBool,
				Description: "Automaticaly create PTR records for the zone.",
				Optional:    true,
				ForceNew:    false,
				Default:     false,
			},
			"notify": {
				Type:         schema.TypeString,
				Description:  "The expected notify behavior (Supported: empty (Inherited), Yes, No, Explicit; Default: empty (Inherited).",
				Optional:     true,
				ForceNew:     false,
				Default:      "",
				ValidateFunc: validation.StringInSlice([]string{"", "yes", "no", "explicit"}, false),
			},
			"also_notify": {
				Type:        schema.TypeList,
				Description: "The list of IP addresses (Format <IP>:<Port>) that will receive zone change notifications in addition to the NS listed in the SOA.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the zone.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to the zone.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourcednszonevalidatetype(v interface{}, _ string) ([]string, []error) {
	switch strings.ToLower(v.(string)) {
	case "master":
		return nil, nil
	default:
		return nil, []error{fmt.Errorf("Unsupported zone type.")}
	}
}

func resourcednszoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Gather required ID(s) from provided information
	siteID, siteErr := ipsiteidbyname(d.Get("space").(string), meta)
	if siteErr != nil {
		// Reporting a failure
		return diag.FromErr(siteErr)
	}

	// Building parameters
	parameters := url.Values{}
	parameters.Add("add_flag", "new_only")
	parameters.Add("dns_name", d.Get("dnsserver").(string))

	// Add dnsview parameter if it is supplied
	// If no view is specified and server has some configured, trigger an error
	if strings.Compare(d.Get("dnsview").(string), "#") != 0 {
		parameters.Add("dnsview_name", d.Get("dnsview").(string))
	} else {
		if dnsserverhasviews(d.Get("dnsserver").(string), meta) {
			return diag.Errorf("Error creating DNS zone: %s, this DNS server has views. Please specify a view name.\n", d.Get("name").(string))
		}
	}

	parameters.Add("dnszone_name", d.Get("name").(string))
	parameters.Add("dnszone_type", strings.ToLower(d.Get("type").(string)))
	parameters.Add("dnszone_site_id", siteID)

	// Building Notify and Also Notify Statements
	parameters.Add("dnszone_notify", strings.ToLower(d.Get("notify").(string)))

	alsoNotifies := ""
	for _, alsoNotify := range toStringArray(d.Get("also_notify").([]interface{})) {
		if match, _ := regexp.MatchString(regexpIPPort, alsoNotify); match == false {
			return diag.Errorf("Only IP:Port format is supported")
		}
		alsoNotifies += strings.Replace(alsoNotify, ":", " port ", 1) + ";"
	}

	if d.Get("notify").(string) == "" || strings.ToLower(d.Get("notify").(string)) == "no" {
		if alsoNotifies != "" {
			return diag.Errorf("Error creating DNS zone: %s (Notify set to 'Inherited' or 'No' but also_notify list is not empty).", strings.ToLower(d.Get("name").(string)))
		}
		parameters.Add("dnszone_also_notify", alsoNotifies)
	} else {
		parameters.Add("dnszone_also_notify", alsoNotifies)
	}

	parameters.Add("dnszone_class_name", d.Get("class").(string))

	// Building class_parameters
	classParameters := urlfromclassparams(d.Get("class_parameters"))
	// Generate class parameter for createptr if required
	if d.Get("createptr").(bool) {
		classParameters.Add("dnsptr", "1")
	} else {
		classParameters.Add("dnsptr", "0")
	}
	parameters.Add("dnszone_class_parameters", classParameters.Encode())

	// Sending the creation request
	resp, body, err := s.Request("post", "rest/dns_zone_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Created DNS zone (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				if errParam, errParamExist := buf[0]["parameters"].(string); errParamExist {
					return diag.Errorf("Unable to create DNS zone: %s (%s - %s)", d.Get("name").(string), errMsg, errParam)
				}
				return diag.Errorf("Unable to create DNS zone: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to create DNS zone: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednszoneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Gather required ID(s) from provided information
	siteID, siteErr := ipsiteidbyname(d.Get("space").(string), meta)
	if siteErr != nil {
		// Reporting a failure
		return diag.FromErr(siteErr)
	}

	// Building parameters
	parameters := url.Values{}
	parameters.Add("dnszone_id", d.Id())
	parameters.Add("add_flag", "edit_only")
	if strings.Compare(d.Get("dnsview").(string), "#") != 0 {
		parameters.Add("dnsview_name", d.Get("dnsview").(string))
	}
	parameters.Add("dnszone_site_id", siteID)

	// Building Notify and Also Notify Statements
	parameters.Add("dnszone_notify", strings.ToLower(d.Get("notify").(string)))

	alsoNotifies := ""
	for _, alsoNotify := range toStringArray(d.Get("also_notify").([]interface{})) {
		if match, _ := regexp.MatchString(regexpIPPort, alsoNotify); match == false {
			return diag.Errorf("Only IP:Port format is supported")
		}
		alsoNotifies += strings.Replace(alsoNotify, ":", " port ", 1) + ";"
	}

	if d.Get("notify").(string) == "" || strings.ToLower(d.Get("notify").(string)) == "no" {
		if alsoNotifies != "" {
			return diag.Errorf("Error updating DNS zone: %s (Notify set to 'Inherited' or 'No' but also_notify list is not empty).", strings.ToLower(d.Get("name").(string)))
		}
		parameters.Add("dnszone_also_notify", alsoNotifies)
	} else {
		parameters.Add("dnszone_also_notify", alsoNotifies)
	}

	parameters.Add("dnszone_class_name", d.Get("class").(string))

	// Building class_parameters
	classParameters := urlfromclassparams(d.Get("class_parameters"))
	// Generate class parameter for createptr if required
	if d.Get("createptr").(bool) {
		classParameters.Add("dnsptr", "1")
	} else {
		classParameters.Add("dnsptr", "0")
	}
	parameters.Add("dnszone_class_parameters", classParameters.Encode())

	// Sending the update request
	resp, body, err := s.Request("put", "rest/dns_zone_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Updated DNS zone (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				if errParam, errParamExist := buf[0]["parameters"].(string); errParamExist {
					return diag.Errorf("Unable to update DNS zone: %s (%s - %s)", d.Get("name").(string), errMsg, errParam)
				}
				return diag.Errorf("Unable to update DNS zone: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to update DNS zone: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednszoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("dnszone_id", d.Id())

	if strings.Compare(d.Get("dnsview").(string), "#") != 0 {
		parameters.Add("dnsview_name", d.Get("dnsview").(string))
	}

	// Sending the deletion request
	resp, body, err := s.Request("delete", "rest/dns_zone_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					return diag.Errorf("Unable to delete DNS zone: %s (%s)", d.Get("name").(string), errMsg)
				}
			}

			return diag.Errorf("Unable to delete DNS zone: %s", d.Get("name").(string))
		}

		// Log deletion
		tflog.Debug(ctx, fmt.Sprintf("Deleted DNS zone (oid): %s\n", d.Id()))

		// Unset local ID
		d.SetId("")

		// Reporting a success
		return nil
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednszoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("dnszone_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/dns_zone_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("dnsserver", buf[0]["dns_name"].(string))
			d.Set("dnsview", buf[0]["dnsview_name"].(string))
			d.Set("name", buf[0]["dnszone_name"].(string))
			d.Set("type", buf[0]["dnszone_type"].(string))

			if buf[0]["dnszone_site_name"].(string) != "#" {
				d.Set("space", buf[0]["dnszone_site_name"].(string))
			} else {
				d.Set("space", "")
			}

			d.Set("notify", strings.ToLower(buf[0]["dnszone_notify"].(string)))
			if buf[0]["dnszone_also_notify"].(string) != "" {
				d.Set("also_notify", toStringArrayInterface(strings.Split(strings.ReplaceAll(strings.TrimSuffix(buf[0]["dnszone_also_notify"].(string), ";"), " port ", ":"), ";")))
			}

			d.Set("class", buf[0]["dnszone_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["dnszone_class_parameters"].(string))
			computedClassParameters := map[string]string{}

			if createptr, createptrExist := retrievedClassParameters["dnsptr"]; createptrExist {
				if createptr[0] == "1" {
					d.Set("createptr", true)
				} else {
					d.Set("createptr", false)
				}
				delete(retrievedClassParameters, "dnsptr")
			}

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
				tflog.Debug(ctx, fmt.Sprintf("Unable to find DNS zone: %s (%s)\n", d.Get("name"), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find DNS zone (oid): %s\n", d.Id()))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("Unable to find DNS zone: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednszoneImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("dnszone_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/dns_zone_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("dnsserver", buf[0]["dns_name"].(string))
			d.Set("dnsview", buf[0]["dnsview_name"].(string))
			d.Set("name", buf[0]["dnszone_name"].(string))
			d.Set("type", buf[0]["dnszone_type"].(string))

			if buf[0]["dnszone_site_name"].(string) != "#" {
				d.Set("space", buf[0]["dnszone_site_name"].(string))
			} else {
				d.Set("space", "")
			}

			d.Set("notify", strings.ToLower(buf[0]["dnszone_notify"].(string)))
			if buf[0]["dnszone_also_notify"].(string) != "" {
				d.Set("also_notify", toStringArrayInterface(strings.Split(strings.ReplaceAll(strings.TrimSuffix(buf[0]["dnszone_also_notify"].(string), ";"), " port ", ":"), ";")))
			}

			d.Set("class", buf[0]["dnszone_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["dnszone_class_parameters"].(string))
			computedClassParameters := map[string]string{}

			if createptr, createptrExist := retrievedClassParameters["dnsptr"]; createptrExist {
				if createptr[0] == "1" {
					d.Set("createptr", true)
				} else {
					d.Set("createptr", false)
				}
				delete(retrievedClassParameters, "dnsptr")
			}

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
				tflog.Debug(ctx, fmt.Sprintf("Unable to import DNS zone (oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import DNS zone (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("SOLIDServer - Unable to find and import DNS zone (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
