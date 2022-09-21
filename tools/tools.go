//go:build tools
// +build tools

package tools

//go:generate go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

import (
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)