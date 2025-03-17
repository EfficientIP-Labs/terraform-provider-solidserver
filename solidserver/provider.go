package solidserver

import (
	"context"
	"net/url"
	"regexp"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:         schema.TypeString,
				Required:     true,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"SOLIDSERVER_HOST", "SOLIDServer_HOST"}, nil),
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "SOLIDServer Hostname or IP address. This can also be specified via the SOLIDSERVER_HOST environment variable",
			},
			"use_token": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"SOLIDSERVER_USE_TOKEN", "SOLIDServer_USE_TOKEN"}, false),
				Description: "SOLIDServer username/password are token/secret. This can also be specified via the SOLIDSERVER_USE_TOKEN environment variable",
			},
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"SOLIDSERVER_USERNAME", "SOLIDServer_USERNAME"}, nil),
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "SOLIDServer API User ID or Token ID. This can also be specified via the SOLIDSERVER_USERNAME environment variable",
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"SOLIDSERVER_PASSWORD", "SOLIDServer_PASSWORD"}, nil),
				ValidateFunc: validation.StringIsNotEmpty,
				Description:  "SOLIDServer API user password or token secret. This can also be specified via the SOLIDSERVER_PASSWORD environment variable",
			},
			"sslverify": {
				Type:        schema.TypeBool,
				Required:    false,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"SOLIDSERVER_SSLVERIFY", "SOLIDServer_SSLVERIFY"}, true),
				Description: "Enable/Disable ssl verify (Default : enabled). This can also be specified via the SOLIDSERVER_SSLVERIFY environment variable",
			},
			"additional_trust_certs_file": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"SOLIDSERVER_ADDITIONALTRUSTCERTSFILE", "SOLIDServer_ADDITIONALTRUSTCERTSFILE"}, nil),
				Description: "PEM formatted file with additional certificates to trust for TLS connection. This can also be specified via the SOLIDSERVER_ADDITIONALTRUSTCERTSFILE environment variable",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Required:    false,
				Optional:    true,
				Description: "API call timeout value in seconds (Default 10s)",
				Default:     10,
			},
			"solidserverversion": {
				Type:         schema.TypeString,
				Required:     false,
				Optional:     true,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"SOLIDSERVER_VERSION", "SOLIDServer_VERSION"}, ""),
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-9]\.[0-9]\.[0-9]((\.[pP]\d+[a-z]?)|[a-z])?)?$`), "Invalid Version Number"),
				Description:  "SOLIDServer Version in case API user does not have admin permissions. This can also be specified via the SOLIDSERVER_VERSION environment variable",
			},
			"proxy_url": {
				Type:             schema.TypeString,
				Required:         false,
				Optional:         true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"SOLIDSERVER_PROXY_URL", "SOLIDServer_PROXY_URL"}, ""),
				Description:      "URL for a proxy to be used for SOLIDServer connectivity. Empty or unspecified means no proxy (direct connectivity). Supported URL schemes are 'http', 'https', and 'socks5'. If the scheme is empty, 'http' is assumed. This can also be specified via the SOLIDSERVER_PROXY_URL environment variable",
				ValidateDiagFunc: validateProxyURLValue,
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"solidserver_ip_space":         dataSourceipspace(),
			"solidserver_ip_subnet":        dataSourceipsubnet(),
			"solidserver_ip_subnet_query":  dataSourceipsubnetquery(),
			"solidserver_ip6_subnet":       dataSourceip6subnet(),
			"solidserver_ip6_subnet_query": dataSourceip6subnetquery(),
			"solidserver_ip_pool":          dataSourceippool(),
			"solidserver_ip6_pool":         dataSourceip6pool(),
			"solidserver_ip_address":       dataSourceipaddress(),
			"solidserver_ip6_address":      dataSourceip6address(),
			"solidserver_ip_ptr":           dataSourceipptr(),
			"solidserver_ip6_ptr":          dataSourceip6ptr(),
			"solidserver_dns_smart":        dataSourcednssmart(),
			"solidserver_dns_server":       dataSourcednsserver(),
			"solidserver_dns_view":         dataSourcednsview(),
			"solidserver_dns_zone":         dataSourcednszone(),
			"solidserver_vlan_domain":      dataSourcevlandomain(),
			"solidserver_vlan_range":       dataSourcevlanrange(),
			"solidserver_vlan":             dataSourcevlan(),
			"solidserver_usergroup":        dataSourceusergroup(),
			"solidserver_cdb":              dataSourcecdb(),
			"solidserver_cdb_data":         dataSourcecdbdata(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"solidserver_ip_space":         resourceipspace(),
			"solidserver_ip_subnet":        resourceipsubnet(),
			"solidserver_ip6_subnet":       resourceip6subnet(),
			"solidserver_ip_pool":          resourceippool(),
			"solidserver_ip6_pool":         resourceip6pool(),
			"solidserver_ip_address":       resourceipaddress(),
			"solidserver_ip6_address":      resourceip6address(),
			"solidserver_ip_alias":         resourceipalias(),
			"solidserver_ip6_alias":        resourceip6alias(),
			"solidserver_ip_mac":           resourceipmac(),
			"solidserver_ip6_mac":          resourceip6mac(),
			"solidserver_device":           resourcedevice(),
			"solidserver_vlan_domain":      resourcevlandomain(),
			"solidserver_vlan_range":       resourcevlanrange(),
			"solidserver_vlan":             resourcevlan(),
			"solidserver_dns_smart":        resourcednssmart(),
			"solidserver_dns_server":       resourcednsserver(),
			"solidserver_dns_view":         resourcednsview(),
			"solidserver_dns_zone":         resourcednszone(),
			"solidserver_dns_forward_zone": resourcednsforwardzone(),
			"solidserver_dns_rr":           resourcednsrr(),
			"solidserver_app_application":  resourceapplication(),
			"solidserver_app_pool":         resourceapplicationpool(),
			"solidserver_app_node":         resourceapplicationnode(),
			"solidserver_user":             resourceuser(),
			"solidserver_usergroup":        resourceusergroup(),
			"solidserver_cdb":              resourcecdb(),
			"solidserver_cdb_data":         resourcecdbdata(),
		},
		ConfigureContextFunc: ProviderConfigure,
	}
}

func ProviderConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	s, err := NewSOLIDserver(
		ctx,
		d.Get("host").(string),
		d.Get("use_token").(bool),
		d.Get("username").(string),
		d.Get("password").(string),
		d.Get("sslverify").(bool),
		d.Get("additional_trust_certs_file").(string),
		d.Get("timeout").(int),
		d.Get("solidserverversion").(string),
		d.Get("proxy_url").(string),
	)
	return s, err
}

func validateProxyURLValue(value interface{}, path cty.Path) diag.Diagnostics {
	proxyURLValue := value.(string)

	// Empty value corresponds to no configured proxy
	if proxyURLValue == "" {
		return nil
	}

	proxyURL, err := url.Parse(proxyURLValue)
	if err != nil {
		return diag.FromErr(path.NewErrorf("invalid url: %w", err))
	}

	if proxyURL.Scheme != "" {
		// Supported schemes are taken from the golang `http.Transport` type Proxy field docs
		// https://pkg.go.dev/net/http#Transport
		validSchemes := map[string]bool{"http": true, "https": true, "socks5": true}
		if _, ok := validSchemes[proxyURL.Scheme]; !ok {
			return diag.FromErr(path.NewErrorf("unsupported proxy url scheme: %s", proxyURL.Scheme))
		}
	}

	return nil
}
