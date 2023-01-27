package datasource_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	tu "github.com/asheliahut/terraform-provider-infisical/testingutils"
)

func TestAccOrganizationsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: tu.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: tu.ProviderConfig + `data "infisical_organizations" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify number of coffees returned
					resource.TestCheckResourceAttr("data.infisical_organizations.test", "organizations.#", "1"),
					// Verify the first coffee to ensure all attributes are set
					resource.TestCheckResourceAttr("data.infisical_organizations.test", "organizations.0.id", ""),
					resource.TestCheckResourceAttr("data.infisical_organizations.test", "organizations.0.name", ""),
				),
			},
		},
	})
}
