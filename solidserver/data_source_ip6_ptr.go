package solidserver

import (
	"context"
	"github.com/MakeNowJust/heredoc/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"math/rand"
	"strconv"
)

func dataSourceip6ptr() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceip6ptrRead,

		Description: heredoc.Doc(`
			IPv6 PTR data-source allows to easily convert an IPv6 address into a DNS PTR format.
		`),

		Schema: map[string]*schema.Schema{
			"address": {
				Type:         schema.TypeString,
				Description:  "The IPv6 address to convert into PTR domain name.",
				ValidateFunc: validation.IsIPAddress,
				Required:     true,
			},
			"dname": {
				Type:        schema.TypeString,
				Description: "The PTR record FQDN associated to the IPv6 address.",
				Computed:    true,
			},
		},
	}
}

func dataSourceip6ptrRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dname := ip6toptr(d.Get("address").(string))

	if dname != "" {
		d.SetId(strconv.Itoa(rand.Intn(1000000)))
		d.Set("dname", dname)
		return nil
	}

	// Reporting a failure
	return diag.Errorf("Unable to convert the following IPv6 address into PTR domain name: %s\n", d.Get("address").(string))
}
