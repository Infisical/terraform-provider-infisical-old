package testingutils

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/asheliahut/terraform-provider-infisical/provider"
)

// providerConfig is a shared configuration to combine with the actual
// test configuration so the Infisical client is properly configured.
// It is also possible to use the INFISICAL_ environment variables instead,
// such as updating the Makefile and running the testing through that tool.

var ProviderConfig string

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can reattach.

var TestAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

func init() {
	ProviderConfig = `
	provider "infisical" {
		api_token = "ak.token"
		host      = "https://infisical.com"
	}
	`

	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"infisical": providerserver.NewProtocol6WithError(provider.New()),
	}
}

func PrintOutput(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		fmt.Println(rs.Primary.Attributes)
	}
	return nil
}
