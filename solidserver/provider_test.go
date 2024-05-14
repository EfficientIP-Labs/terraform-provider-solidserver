package solidserver

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testProviders map[string]terraform.ResourceProvider
var testProvider *schema.Provider

func testAccPreCheck(t *testing.T) {
	log.Printf("[DEBUG] - testPreCheck\n")
}

func init() {
	if os.Getenv("SOLIDServer_HOST") == "" {
		fmt.Println("[ERROR] use SOLIDServer_HOST as SOLIDserver target")
		return
	}

	if os.Getenv("SOLIDServer_USERNAME") == "" {
		fmt.Println("[ERROR] use SOLIDServer_USERNAME as SOLIDserver user for API")
		return
	}

	if os.Getenv("SOLIDServer_PASSWORD") == "" {
		fmt.Println("[ERROR] use SOLIDServer_PASSWORD as SOLIDserver password for API")
		return
	}

	if os.Getenv("SOLIDServer_SSLVERIFY") == "" {
		fmt.Println("[WARN] use SOLIDServer_SSLVERIFY=false to bypass certificate validation")
	}

	testProvider = Provider().(*schema.Provider)
	testProviders = map[string]terraform.ResourceProvider{
		"solidserver": testProvider,
	}
}

func TestValidateProxyURLValue(t *testing.T) {

	type testCase struct {
		URL   string
		IsErr bool
	}

	testCases := map[string]testCase{
		"empty_proxy_url": {
			URL: "",
		},
		"no_scheme_url": {
			URL: "proxy.example.com",
		},
		"http_url": {
			URL: "http://proxy.example.com",
		},
		"https_url": {
			URL: "https://proxy.example.com",
		},
		"socks5_url": {
			URL: "socks5://proxy.example.com",
		},
		"invalid_url": {
			URL:   "invalid url:",
			IsErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := validateProxyURLValue(tc.URL, cty.GetAttrPath("proxy_url"))

			if tc.IsErr {
				if !result.HasError() {
					t.Errorf("expected error")
				}
			} else {
				if result.HasError() {
					t.Errorf("unexpected error: %+v", result)
				}
			}
		})
	}
}
