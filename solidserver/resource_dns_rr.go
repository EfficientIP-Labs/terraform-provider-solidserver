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
	"strings"
)

func resourcednsrr() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcednsrrCreate,
		ReadContext:   resourcednsrrRead,
		UpdateContext: resourcednsrrUpdate,
		DeleteContext: resourcednsrrDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcednsrrImportState,
		},

		Description: heredoc.Doc(`
			DNS RR resource allows to create and manage DNS resource records of type A, AAAA, PTR, CNAME, DNAME, NS.
		`),

		Schema: map[string]*schema.Schema{
			"dnsserver": {
				Type:        schema.TypeString,
				Description: "The managed SMART DNS server name, or DNS server name hosting the RR's zone.",
				Required:    true,
				ForceNew:    true,
			},
			"dnsview": {
				Type:        schema.TypeString,
				Description: "The View name of the RR to create.",
				Optional:    true,
				ForceNew:    true,
				Default:     "",
			},
			"dnszone": {
				Type:        schema.TypeString,
				Description: "The Zone name of the RR to create.",
				Optional:    true,
				ForceNew:    true,
				Default:     "",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The Fully Qualified Domain Name of the RR to create.",
				Required:    true,
				ForceNew:    true,
			},
			"type": {
				Type:         schema.TypeString,
				Description:  "The type of the RR to create (Supported: A, AAAA, PTR, CNAME, DNAME and NS).",
				ValidateFunc: resourcednsrrvalidatetype,
				Required:     true,
				ForceNew:     true,
			},
			"value": {
				Type:             schema.TypeString,
				Description:      "The value od the RR to create.",
				Computed:         false,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: resourcediffsuppressIPv6Format,
			},
			"ttl": {
				Type:        schema.TypeInt,
				Description: "The DNS Time To Live of the RR to create.",
				Optional:    true,
				Default:     3600,
			},
			"class": {
				Type:        schema.TypeString,
				Description: "The class associated to the DNS view.",
				Optional:    true,
				ForceNew:    false,
				Default:     "",
			},
			"class_parameters": {
				Type:        schema.TypeMap,
				Description: "The class parameters associated to the view.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourcednsrrvalidatetype(v interface{}, _ string) ([]string, []error) {
	switch strings.ToUpper(v.(string)) {
	case "A":
		return nil, nil
	case "AAAA":
		return nil, nil
	case "PTR":
		return nil, nil
	case "CNAME":
		return nil, nil
	case "DNAME":
		return nil, nil
	case "TXT":
		return nil, nil
	case "NS":
		return nil, nil
	default:
		return nil, []error{fmt.Errorf("Unsupported RR type.")}
	}
}

func resourcednsrrCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("add_flag", "new_only")
	parameters.Add("dns_name", d.Get("dnsserver").(string))
	parameters.Add("rr_name", d.Get("name").(string))
	parameters.Add("rr_type", strings.ToUpper(d.Get("type").(string)))
	parameters.Add("value1", d.Get("value").(string))
	parameters.Add("rr_ttl", strconv.Itoa(d.Get("ttl").(int)))

	// Add dnsview parameter if it is supplied
	if len(d.Get("dnsview").(string)) != 0 {
		parameters.Add("dnsview_name", strings.ToLower(d.Get("dnsview").(string)))
	}

	// Add dnszone parameter if it is supplied
	if len(d.Get("dnszone").(string)) != 0 {
		parameters.Add("dnszone_name", strings.ToLower(d.Get("dnszone").(string)))
	}

	if s.Version < 800 {
		tflog.Info(ctx, fmt.Sprintf("RR class parameters are not supported in SOLIDserver Version (%i)", s.Version))
	} else {
		parameters.Add("rr_class_name", d.Get("class").(string))
		parameters.Add("rr_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())
	}

	// Sending the creation request
	resp, body, err := s.Request("post", "rest/dns_rr_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Created RR (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to create RR: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to create RR: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednsrrUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("rr_id", d.Id())
	parameters.Add("add_flag", "edit_only")
	parameters.Add("dns_name", d.Get("dnsserver").(string))
	parameters.Add("rr_name", d.Get("name").(string))
	parameters.Add("rr_type", strings.ToUpper(d.Get("type").(string)))
	parameters.Add("value1", d.Get("value").(string))
	parameters.Add("rr_ttl", strconv.Itoa(d.Get("ttl").(int)))

	// Add dnsview parameter if it is supplied
	if len(d.Get("dnsview").(string)) != 0 {
		parameters.Add("dnsview_name", strings.ToLower(d.Get("dnsview").(string)))
	}

	// Add dnszone parameter if it is supplied
	if len(d.Get("dnszone").(string)) != 0 {
		parameters.Add("dnszone_name", strings.ToLower(d.Get("dnszone").(string)))
	}

	if s.Version < 800 {
		tflog.Info(ctx, fmt.Sprintf("RR class parameters are not supported in SOLIDserver Version (%i)", s.Version))
	} else {
		parameters.Add("rr_class_name", d.Get("class").(string))
		parameters.Add("rr_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())
	}

	// Sending the update request
	resp, body, err := s.Request("put", "rest/dns_rr_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Updated RR (oid): %s\n", oid))
				d.SetId(oid)
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				return diag.Errorf("Unable to update RR: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to update RR: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednsrrDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("rr_id", d.Id())

	// Add dnsview parameter if it is supplied
	if len(d.Get("dnsview").(string)) != 0 {
		parameters.Add("dnsview_name", strings.ToLower(d.Get("dnsview").(string)))
	}

	// Sending the deletion request
	resp, body, err := s.Request("delete", "rest/dns_rr_delete", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode != 200 && resp.StatusCode != 204 {
			// Reporting a failure
			if len(buf) > 0 {
				if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
					return diag.Errorf("Unable to delete RR: %s (%s)", d.Get("name").(string), errMsg)
				}
			}

			return diag.Errorf("Unable to delete RR: %s", d.Get("name").(string))
		}

		// Log deletion
		tflog.Debug(ctx, fmt.Sprintf("Deleted RR (oid): %s\n", d.Id()))

		// Unset local ID
		d.SetId("")

		// Reporting a success
		return nil
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednsrrRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}

	// Sending the read request
	// We do not rely on the ID that may change due to DNS behavior
	whereClause := "dns_name='" + d.Get("dnsserver").(string) + "' AND rr_full_name='" + d.Get("name").(string) + "' AND rr_type='" + strings.ToUpper(d.Get("type").(string))

	if strings.ToUpper(d.Get("type").(string)) == "AAAA" {
		value := shortip6tolongip6(d.Get("value").(string))
		tflog.Debug(ctx, fmt.Sprintf("Using Expanded IPv6 format: %s\n", value))
		whereClause += "' AND value1='" + value + "' "
	} else {
		whereClause += "' AND value1='" + d.Get("value").(string) + "' "
	}

	// Attempt to hande changing RR IDs
	if len(d.Get("dnsview").(string)) != 0 {
		whereClause += "AND dnsview_name='" + d.Get("dnsview").(string) + "' "
	} else {
		whereClause += "AND dnsview_name='#' "
	}

	// Add dnszone parameter if it is supplied
	if len(d.Get("dnszone").(string)) != 0 {
		whereClause += "AND dnszone_name='" + d.Get("dnszone").(string) + "' "
	}

	parameters.Add("WHERE", whereClause)
	resp, body, err := s.Request("get", "rest/dns_rr_list", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			if oid, oidExist := buf[0]["rr_id"].(string); oidExist {
				d.SetId(oid)
			}

			ttl, _ := strconv.Atoi(buf[0]["ttl"].(string))

			d.Set("dnsserver", buf[0]["dns_name"].(string))
			d.Set("name", buf[0]["rr_full_name"].(string))
			d.Set("type", buf[0]["rr_type"].(string))

			if strings.ToUpper(buf[0]["rr_type"].(string)) == "AAAA" {
				d.Set("value", longip6toshortip6(buf[0]["value1"].(string)))
			} else {
				d.Set("value", buf[0]["value1"].(string))
			}

			d.Set("ttl", ttl)

			if buf[0]["dnsview_name"].(string) != "#" {
				d.Set("dnsview", buf[0]["dnsview_name"].(string))
			}

			if s.Version < 800 {
				tflog.Info(ctx, fmt.Sprintf("RR class parameters are not supported in SOLIDserver Version (%i)", s.Version))
			} else {
				d.Set("class", buf[0]["rr_class_name"].(string))

				// Updating local class_parameters
				currentClassParameters := d.Get("class_parameters").(map[string]interface{})
				retrievedClassParameters, _ := url.ParseQuery(buf[0]["rr_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to find RR: %s (%s)\n", d.Get("name"), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find RR (oid): %s\n", d.Id()))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("SOLIDServer - Unable to find RR: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednsrrImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("rr_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/dns_rr_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			ttl, _ := strconv.Atoi(buf[0]["ttl"].(string))

			d.Set("dnsserver", buf[0]["dns_name"].(string))
			d.Set("name", buf[0]["rr_full_name"].(string))
			d.Set("type", buf[0]["rr_type"].(string))

			if strings.ToUpper(buf[0]["rr_type"].(string)) == "AAAA" {
				d.Set("value", longip6toshortip6(buf[0]["value1"].(string)))
			} else {
				d.Set("value", buf[0]["value1"].(string))
			}

			d.Set("ttl", ttl)

			if buf[0]["dnsview_name"].(string) != "#" {
				d.Set("dnsview", buf[0]["dnsview_name"].(string))
			}

			if s.Version < 800 {
				tflog.Info(ctx, fmt.Sprintf("RR class parameters are not supported in SOLIDserver Version (%i)", s.Version))
			} else {
				d.Set("class", buf[0]["rr_class_name"].(string))

				// Updating local class_parameters
				currentClassParameters := d.Get("class_parameters").(map[string]interface{})
				retrievedClassParameters, _ := url.ParseQuery(buf[0]["rr_class_parameters"].(string))
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
				// Log the error
				tflog.Debug(ctx, fmt.Sprintf("Unable to import RR (oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import RR (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("Unable to find and import RR (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
