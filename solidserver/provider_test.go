package solidserver

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
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
