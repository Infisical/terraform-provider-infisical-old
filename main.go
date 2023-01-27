package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/asheliahut/terraform-provider-infisical/provider"
)

// Provider http client generation.
//go:generate oapi-codegen --package=client -generate=client,types -o ./client/infisical.gen.go https://raw.githubusercontent.com/Infisical/infisical/main/docs/spec.yaml

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name infisical
func main() {
	providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/infisical/infisical",
	})
}
