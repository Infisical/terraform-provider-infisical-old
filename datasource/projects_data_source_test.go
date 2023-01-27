package datasource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	tu "github.com/asheliahut/terraform-provider-infisical/testingutils"
)

var testConfig = `
data "infisical_organizations" "test" {}

data "infisical_projects" "test" {
    organization_id = data.infisical_organizations.test.organizations.0.id
}
`

func TestAccProjectsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tu.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: tu.ProviderConfig + testConfig,

				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of coffees returned
					resource.TestCheckResourceAttr("data.infisical_projects.test", "projects.#", "1"),
					// Verify top level data
					resource.TestCheckResourceAttr("data.infisical_projects.test", "organization_id", ""),
					// Verify the first coffee to ensure all attributes are set
					resource.TestCheckResourceAttr("data.infisical_projects.test", "projects.0.id", ""),
					resource.TestCheckResourceAttr("data.infisical_projects.test", "projects.0.name", ""),
				),
			},
		},
	})
}
