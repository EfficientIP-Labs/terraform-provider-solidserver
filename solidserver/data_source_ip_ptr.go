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

func dataSourceipptr() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceipptrRead,

		Description: heredoc.Doc(`
			IP PTR data-source allows to easily convert an IPv4 address into a DNS PTR format.
		`),

		Schema: map[string]*schema.Schema{
			"address": {
				Type:         schema.TypeString,
				Description:  "The IP address to convert into PTR domain name.",
				ValidateFunc: validation.IsIPAddress,
				Required:     true,
			},
			"dname": {
				Type:        schema.TypeString,
				Description: "The PTR record FQDN associated to the IP address.",
				Computed:    true,
			},
		},
	}
}

func dataSourceipptrRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	dname := iptoptr(d.Get("address").(string))

	if dname != "" {
		d.SetId(strconv.Itoa(rand.Intn(1000000)))
		d.Set("dname", dname)
		return nil
	}

	// Reporting a failure
	return diag.Errorf("Unable to convert the following IP address into PTR domain name: %s\n", d.Get("address").(string))
}
