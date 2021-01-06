package iacprovider

import (
	"reflect"

	tfv12 "github.com/accurics/terrascan/pkg/iac-providers/terraform/v12"
	tfv14 "github.com/accurics/terrascan/pkg/iac-providers/terraform/v14"
)

// terraform specific constants
const (
	terraform                 supportedIacType    = "terraform"
	terraformV12              supportedIacVersion = "v12"
	terraformV14              supportedIacVersion = "v14"
	terraformDefaultVersion                       = terraformV12
	terraformDefaultVersion14                     = terraformV14
)

// register terraform as an IaC provider with terrascan
func init() {
	// register iac provider
	RegisterIacProvider(terraform, terraformV12, terraformDefaultVersion, reflect.TypeOf(tfv12.TfV12{}))
	RegisterIacProvider(terraform, terraformV14, terraformDefaultVersion14, reflect.TypeOf(tfv14.TfV14{}))
}
