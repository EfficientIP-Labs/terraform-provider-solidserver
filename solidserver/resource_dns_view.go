package solidserver

import (
	//"encoding/hex"
	"context"
	"encoding/json"
	"fmt"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func resourcednsview() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcednsviewCreate,
		ReadContext:   resourcednsviewRead,
		UpdateContext: resourcednsviewUpdate,
		DeleteContext: resourcednsviewDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcednsviewImportState,
		},

		Description: heredoc.Doc(`
			DNS View resource allows to create and configure DNS views.
			View(s) are virtual containers mostly used to implement DNS split horizon
			providing different answers depending on matching criteria.
		`),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the DNS view to create.",
				//DiffSuppressFunc: resourcediffsuppresscase,
				// See https://developer.hashicorp.com/terraform/plugin/sdkv2/resources/customizing-differences
				Required: true,
				ForceNew: true,
			},
			"dnsserver": {
				Type:             schema.TypeString,
				Description:      "The name of DNS server or DNS SMART hosting the DNS view to create.",
				DiffSuppressFunc: resourcediffsuppresscase,
				Required:         true,
				ForceNew:         true,
			},
			"order": {
				Type:        schema.TypeInt,
				Description: "The level of the DNS view, where 0 represents the highest level in the views hierarchy.",
				Computed:    true,
				ForceNew:    false,
			},
			"recursion": {
				Type:        schema.TypeBool,
				Description: "The recursion mode of the DNS view (Default: true).",
				Optional:    true,
				Default:     false,
			},
			"forward": {
				Type:         schema.TypeString,
				Description:  "The forwarding mode of the DNS SMART (Supported: none, first, only; Default: none).",
				ValidateFunc: validation.StringInSlice([]string{"none", "first", "only"}, false),
				Optional:     true,
				Default:      "none",
			},
			"forwarders": {
				Type:        schema.TypeList,
				Description: "The IP address list of the forwarder(s) configured to configure on the DNS SMART.",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// ACL(s)
			// Views and Servers/SMARTs
			"allow_transfer": {
				Type:        schema.TypeList,
				Description: "A list of network prefixes allowed to query the view for zone transfert (named ACL(s) are not supported using this provider).  Use '!' to negate an entry.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"allow_query": {
				Type:        schema.TypeList,
				Description: "A list of network prefixes allowed to query the view (named ACL(s) are not supported using this provider).  Use '!' to negate an entry.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"allow_recursion": {
				Type:        schema.TypeList,
				Description: "A list of network prefixes allowed to query the view for recursion (named ACL(s) are not supported using this provider).  Use '!' to negate an entry.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// Views Only
			"match_clients": {
				Type:        schema.TypeList,
				Description: "A list of network prefixes used to match the clients of the view (named ACL(s) are not supported using this provider).  Use '!' to negate an entry.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"match_to": {
				Type:        schema.TypeList,
				Description: "A list of network prefixes used to match the traffic to the view (named ACL(s) are not supported using this provider).  Use '!' to negate an entry.",
				Optional:    true,
				ForceNew:    false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("name", func(ctx context.Context, old, new, meta any) error {
				if strings.ToLower(new.(string)) != new.(string) && strings.ToLower(old.(string)) == old.(string) && strings.ToLower(new.(string)) == strings.ToLower(old.(string)) {
					return fmt.Errorf("View name contains upper case characters (%s), remote name matches but is lower case (%s). Consider fixing your .tf file(s).", new.(string), old.(string))
				}
				return nil
			}),
		),
	}
}

func resourcednsviewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("add_flag", "new_only")
	parameters.Add("dnsview_name", d.Get("name").(string))
	parameters.Add("dns_name", strings.ToLower(d.Get("dnsserver").(string)))

	// Configure recursion
	if d.Get("recursion").(bool) {
		parameters.Add("dnsview_recursion", "yes")
	} else {
		parameters.Add("dnsview_recursion", "no")
	}

	// Only look for network prefixes, acl(s) names will be ignored during the sync process with SOLIDserver
	// Building allow_transfer ACL
	allowTransfers := ""
	for _, allowTransfer := range toStringArray(d.Get("allow_transfer").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, allowTransfer); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view's allow_transfer parameter")
		}
		allowTransfers += allowTransfer + ";"
	}
	parameters.Add("dnsview_allow_transfer", allowTransfers)

	// Building allow_query ACL
	allowQueries := ""
	for _, allowQuery := range toStringArray(d.Get("allow_query").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, allowQuery); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view's allow_query parameter")
		}
		allowQueries += allowQuery + ";"
	}
	parameters.Add("dnsview_allow_query", allowQueries)

	// Building allow_recursion ACL
	allowRecursions := ""
	for _, allowRecursion := range toStringArray(d.Get("allow_recursion").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, allowRecursion); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view's allow_recursion parameter")
		}
		allowRecursions += allowRecursion + ";"
	}
	parameters.Add("dnsview_allow_recursion", allowRecursions)

	// Building match_clients ACL
	matchClients := ""
	for _, matchClient := range toStringArray(d.Get("match_clients").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, matchClient); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view's match_clients parameter")
		}
		matchClients += matchClient + ";"
	}
	parameters.Add("dnsview_match_clients", matchClients)

	// Building match_to ACL
	matchTos := ""
	for _, matchTo := range toStringArray(d.Get("match_to").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, matchTo); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view match_to parameter")
		}
		matchTos += matchTo + ";"
	}
	parameters.Add("dnsview_match_to", matchTos)

	parameters.Add("dnsview_class_name", d.Get("class").(string))
	parameters.Add("dnsview_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	// Sending creation request
	resp, body, err := s.Request("post", "rest/dns_view_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Created DNS view (oid): %s\n", oid))
				d.SetId(oid)

				// Building forward mode and forward list
				fwdList := ""
				for _, fwd := range toStringArray(d.Get("forwarders").([]interface{})) {
					fwdList += fwd + ";"
				}

				if d.Get("forward").(string) == "none" {
					if fwdList != "" {
						return diag.Errorf("Error creating DNS view: %s (Forward mode set to 'none' but forwarders list is not empty).", d.Get("name").(string))
					}
					// NOT required at creation time - dnsparamunset(d.Get("dnsserver").(string), oid, "forward", meta)
					dnsparamset(d.Get("dnsserver").(string), oid, "forwarders", "", meta)
				} else {
					dnsparamset(d.Get("dnsserver").(string), oid, "forward", strings.ToLower(d.Get("forward").(string)), meta)
					dnsparamset(d.Get("dnsserver").(string), oid, "forwarders", fwdList, meta)
				}

				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				if errParam, errParamExist := buf[0]["parameters"].(string); errParamExist {
					return diag.Errorf("Unable to create DNS view: %s (%s - %s)", d.Get("name").(string), errMsg, errParam)
				}
				return diag.Errorf("Unable to create DNS view: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to create DNS view: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednsviewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("dnsview_id", d.Id())
	parameters.Add("dns_name", strings.ToLower(d.Get("dnsserver").(string)))
	parameters.Add("dnsview_name", d.Get("name").(string))
	parameters.Add("add_flag", "edit_only")

	// Configure recursion
	if d.Get("recursion").(bool) {
		parameters.Add("dnsview_recursion", "yes")
	} else {
		parameters.Add("dnsview_recursion", "no")
	}

	// Only look for network prefixes, acl(s) names will be ignored during the sync process with SOLIDserver
	// Building allow_transfer ACL
	allowTransfers := ""
	for _, allowTransfer := range toStringArray(d.Get("allow_transfer").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, allowTransfer); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view's allow_transfer parameter")
		}
		allowTransfers += allowTransfer + ";"
	}
	parameters.Add("dnsview_allow_transfer", allowTransfers)

	// Building allow_query ACL
	allowQueries := ""
	for _, allowQuery := range toStringArray(d.Get("allow_query").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, allowQuery); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view's allow_query parameter")
		}
		allowQueries += allowQuery + ";"
	}
	parameters.Add("dnsview_allow_query", allowQueries)

	// Building allow_recursion ACL
	allowRecursions := ""
	for _, allowRecursion := range toStringArray(d.Get("allow_recursion").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, allowRecursion); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view's allow_recursion parameter")
		}
		allowRecursions += allowRecursion + ";"
	}
	parameters.Add("dnsview_allow_recursion", allowRecursions)

	// Building match_clients ACL
	matchClients := ""
	for _, matchClient := range toStringArray(d.Get("match_clients").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, matchClient); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view's match_clients parameter")
		}
		matchClients += matchClient + ";"
	}
	parameters.Add("dnsview_match_clients", matchClients)

	// Building match_to ACL
	matchTos := ""
	for _, matchTo := range toStringArray(d.Get("match_to").([]interface{})) {
		if match, _ := regexp.MatchString(regexpNetworkAcl, matchTo); match == false {
			return diag.Errorf("Only network prefixes are supported for DNS view match_to parameter")
		}
		matchTos += matchTo + ";"
	}
	parameters.Add("dnsview_match_to", matchTos)

	parameters.Add("dnsview_class_name", d.Get("class").(string))
	parameters.Add("dnsview_class_parameters", urlfromclassparams(d.Get("class_parameters")).Encode())

	// Sending the update request
	resp, body, err := s.Request("put", "rest/dns_view_add", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if (resp.StatusCode == 200 || resp.StatusCode == 201) && len(buf) > 0 {
			if oid, oidExist := buf[0]["ret_oid"].(string); oidExist {
				tflog.Debug(ctx, fmt.Sprintf("Updated DNS view (oid): %s\n", oid))
				d.SetId(oid)

				// Building forward mode and forward list
				fwdList := ""
				for _, fwd := range toStringArray(d.Get("forwarders").([]interface{})) {
					fwdList += fwd + ";"
				}

				if d.Get("forward").(string) == "none" {
					if fwdList != "" {
						return diag.Errorf("Error creating DNS view: %s (Forward mode set to 'none' but forwarders list is not empty).", d.Get("name").(string))
					}
					dnsparamunset(d.Get("dnsserver").(string), oid, "forward", meta)
					dnsparamset(d.Get("dnsserver").(string), oid, "forwarders", "", meta)
				} else {
					dnsparamset(d.Get("dnsserver").(string), oid, "forward", strings.ToLower(d.Get("forward").(string)), meta)
					dnsparamset(d.Get("dnsserver").(string), oid, "forwarders", fwdList, meta)
				}
				return nil
			}
		}

		// Reporting a failure
		if len(buf) > 0 {
			if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
				if errParam, errParamExist := buf[0]["parameters"].(string); errParamExist {
					return diag.Errorf("Unable to update DNS view: %s (%s - %s)", d.Get("name").(string), errMsg, errParam)
				}
				return diag.Errorf("Unable to update DNS view: %s (%s)", d.Get("name").(string), errMsg)
			}
		}

		return diag.Errorf("Unable to update DNS view: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednsviewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	for i := 0; i < 3; i++ {
		// Building parameters
		parameters := url.Values{}
		parameters.Add("dnsview_id", d.Id())

		// Sending the deletion request
		resp, body, err := s.Request("delete", "rest/dns_view_delete", &parameters)

		if err == nil {
			var buf [](map[string]interface{})
			json.Unmarshal([]byte(body), &buf)

			// Checking the answer
			if resp.StatusCode == 200 || resp.StatusCode == 204 {
				// Log deletion
				tflog.Debug(ctx, fmt.Sprintf("Deleted DNS view (oid): %s\n", d.Id()))

				// Unset local ID
				d.SetId("")

				// Reporting a success
				return nil
			} else {
				// Logging a failure
				if len(buf) > 0 {
					if errMsg, errExist := buf[0]["errmsg"].(string); errExist {
						tflog.Debug(ctx, fmt.Sprintf("Unable to delete DNS view: %s (%s)", d.Get("name").(string), errMsg))
					}
				} else {
					tflog.Debug(ctx, fmt.Sprintf("Unable to delete DNS view: %s", d.Get("name").(string)))
				}
				time.Sleep(time.Duration(8 * time.Second))
			}
		} else {
			// Reporting a failure
			return diag.FromErr(err)
		}
	}

	// Reporting a failure
	return diag.Errorf("Unable to delete DNS view: Too many unsuccessful deletion attempts")
}

func resourcednsviewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("dnsview_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/dns_view_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("name", buf[0]["dnsview_name"].(string))
			d.Set("dnsserver", buf[0]["dns_name"].(string))

			viewOrder, _ := strconv.Atoi(buf[0]["dnsview_order"].(string))
			d.Set("order", viewOrder)

			// Updating recursion mode
			if buf[0]["dnsview_recursion"].(string) == "yes" {
				d.Set("recursion", true)
			} else {
				d.Set("recursion", false)
			}

			// Updating forward mode
			forward, forwardErr := dnsparamget(buf[0]["dns_name"].(string), d.Id(), "forward", meta)
			if forwardErr == nil {
				if forward == "" {
					d.Set("forward", "none")
				} else {
					d.Set("forward", strings.ToLower(forward))
				}
			} else {
				tflog.Debug(ctx, fmt.Sprintf("Unable to DNS view's forward mode (oid): %s\n", d.Id()))
				d.Set("forward", "none")
			}

			// Updating forwarder information
			forwarders, forwardersErr := dnsparamget(buf[0]["dns_name"].(string), d.Id(), "forwarders", meta)
			if forwardersErr == nil {
				if forwarders != "" {
					d.Set("forwarders", toStringArrayInterface(strings.Split(strings.TrimSuffix(forwarders, ";"), ";")))
				} else {
					d.Set("forwarders", make([]string, 0))
				}
			} else {
				tflog.Debug(ctx, fmt.Sprintf("Unable to DNS view's forwarders list (oid): %s\n", d.Id()))
				d.Set("forwarders", make([]string, 0))
			}

			// Only look for network prefixes, acl(s) names will be ignored during the sync process with SOLIDserver
			// Building allow_transfer ACL
			if buf[0]["dnsview_allow_transfer"].(string) != "" {
				allowTransfers := []string{}
				for _, allowTransfer := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_allow_transfer"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, allowTransfer.(string)); match == true {
						allowTransfers = append(allowTransfers, allowTransfer.(string))
					}
				}
				d.Set("allow_transfer", allowTransfers)
			}

			// Building allow_query ACL
			if buf[0]["dnsview_allow_query"].(string) != "" {
				allowQueries := []string{}
				for _, allowQuery := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_allow_query"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, allowQuery.(string)); match == true {
						allowQueries = append(allowQueries, allowQuery.(string))
					}
				}
				d.Set("allow_query", allowQueries)
			}

			// Building allow_recursion ACL
			if buf[0]["dnsview_allow_recursion"].(string) != "" {
				allowRecursions := []string{}
				for _, allowRecursion := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_allow_recursion"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, allowRecursion.(string)); match == true {
						allowRecursions = append(allowRecursions, allowRecursion.(string))
					}
				}
				d.Set("allow_recursion", allowRecursions)
			}

			// Updating ACL information
			if buf[0]["dnsview_match_clients"].(string) != "" {
				matchClients := []string{}
				for _, matchClient := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_match_clients"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, matchClient.(string)); match == true {
						matchClients = append(matchClients, matchClient.(string))
					}
				}
				d.Set("match_clients", matchClients)
			}

			if buf[0]["dnsview_match_to"].(string) != "" {
				matchTos := []string{}
				for _, matchTo := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_match_to"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, matchTo.(string)); match == true {
						matchTos = append(matchTos, matchTo.(string))
					}
				}
				d.Set("match_to", matchTos)
			}

			d.Set("class", buf[0]["dnsview_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["dnsview_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to find DNS view: %s (%s)\n", d.Get("name").(string), errMsg))
			}
		} else {
			// Log the error
			tflog.Debug(ctx, fmt.Sprintf("Unable to find DNS view (oid): %s\n", d.Id()))
		}

		// Do not unset the local ID to avoid inconsistency

		// Reporting a failure
		return diag.Errorf("Unable to find DNS view: %s\n", d.Get("name").(string))
	}

	// Reporting a failure
	return diag.FromErr(err)
}

func resourcednsviewImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	s := meta.(*SOLIDserver)

	// Building parameters
	parameters := url.Values{}
	parameters.Add("dnsview_id", d.Id())

	// Sending the read request
	resp, body, err := s.Request("get", "rest/dns_view_info", &parameters)

	if err == nil {
		var buf [](map[string]interface{})
		json.Unmarshal([]byte(body), &buf)

		// Checking the answer
		if resp.StatusCode == 200 && len(buf) > 0 {
			d.Set("name", buf[0]["dnsview_name"].(string))
			d.Set("dnsserver", buf[0]["dns_name"].(string))

			viewOrder, _ := strconv.Atoi(buf[0]["dnsview_order"].(string))
			d.Set("order", viewOrder)

			// Updating recursion mode
			if buf[0]["dnsview_recursion"].(string) == "yes" {
				d.Set("recursion", true)
			} else {
				d.Set("recursion", false)
			}

			// Updating forward mode
			forward, forwardErr := dnsparamget(buf[0]["dns_name"].(string), d.Id(), "forward", meta)
			if forwardErr == nil {
				if forward == "" {
					d.Set("forward", "none")
				} else {
					d.Set("forward", strings.ToLower(forward))
				}
			} else {
				tflog.Debug(ctx, fmt.Sprintf("Unable to DNS view's forward mode (oid): %s\n", d.Id()))
				d.Set("forward", "none")
			}

			// Updating forwarder information
			forwarders, forwardersErr := dnsparamget(buf[0]["dns_name"].(string), d.Id(), "forwarders", meta)
			if forwardersErr == nil {
				if forwarders != "" {
					d.Set("forwarders", toStringArrayInterface(strings.Split(strings.TrimSuffix(forwarders, ";"), ";")))
				} else {
					d.Set("forwarders", make([]string, 0))
				}
			} else {
				tflog.Debug(ctx, fmt.Sprintf("Unable to DNS view's forwarders list (oid): %s\n", d.Id()))
				d.Set("forwarders", make([]string, 0))
			}

			// Only look for network prefixes, acl(s) names will be ignored during the sync process with SOLIDserver
			// Building allow_transfer ACL
			if buf[0]["dnsview_allow_transfer"].(string) != "" {
				allowTransfers := []string{}
				for _, allowTransfer := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_allow_transfer"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, allowTransfer.(string)); match == true {
						allowTransfers = append(allowTransfers, allowTransfer.(string))
					}
				}
				d.Set("allow_transfer", allowTransfers)
			}

			// Building allow_query ACL
			if buf[0]["dnsview_allow_query"].(string) != "" {
				allowQueries := []string{}
				for _, allowQuery := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_allow_query"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, allowQuery.(string)); match == true {
						allowQueries = append(allowQueries, allowQuery.(string))
					}
				}
				d.Set("allow_query", allowQueries)
			}

			// Building allow_recursion ACL
			if buf[0]["dnsview_allow_recursion"].(string) != "" {
				allowRecursions := []string{}
				for _, allowRecursion := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_allow_recursion"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, allowRecursion.(string)); match == true {
						allowRecursions = append(allowRecursions, allowRecursion.(string))
					}
				}
				d.Set("allow_recursion", allowRecursions)
			}

			// Updating ACL information
			if buf[0]["dnsview_match_clients"].(string) != "" {
				matchClients := []string{}
				for _, matchClient := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_match_clients"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, matchClient.(string)); match == true {
						matchClients = append(matchClients, matchClient.(string))
					}
				}
				d.Set("match_clients", matchClients)
			}

			if buf[0]["dnsview_match_to"].(string) != "" {
				matchTos := []string{}
				for _, matchTo := range toStringArrayInterface(strings.Split(strings.TrimSuffix(buf[0]["dnsview_match_to"].(string), ";"), ";")) {
					if match, _ := regexp.MatchString(regexpNetworkAcl, matchTo.(string)); match == true {
						matchTos = append(matchTos, matchTo.(string))
					}
				}
				d.Set("match_to", matchTos)
			}

			d.Set("class", buf[0]["dnsview_class_name"].(string))

			// Updating local class_parameters
			currentClassParameters := d.Get("class_parameters").(map[string]interface{})
			retrievedClassParameters, _ := url.ParseQuery(buf[0]["dnsview_class_parameters"].(string))
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
				tflog.Debug(ctx, fmt.Sprintf("Unable to import DNS view (oid): %s (%s)\n", d.Id(), errMsg))
			}
		} else {
			tflog.Debug(ctx, fmt.Sprintf("Unable to find and import DNS view (oid): %s\n", d.Id()))
		}

		// Reporting a failure
		return nil, fmt.Errorf("SOLIDServer - Unable to find and import DNS view (oid): %s\n", d.Id())
	}

	// Reporting a failure
	return nil, err
}
