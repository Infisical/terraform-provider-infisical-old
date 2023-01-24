package infisical

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	ic "github.com/asheliahut/terraform-provider-infisical/client"
)

func dataSourceOrganizations() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOrganizationsRead,
		Schema: map[string]*schema.Schema{
			"organizations": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"createdAt": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"updatedAt": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"__v": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceOrganizationsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*ic.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Call the API
	r, err := c.GetApiV2UsersMeOrganizations(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	defer r.Body.Close()

	// Convert the response to a slice of maps
	organizationsSlice := make([]map[string]interface{}, 0)
	err = json.NewDecoder(r.Body).Decode(&organizationsSlice)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("organizations", organizationsSlice); err != nil {
		return diag.FromErr(err)
	}

	// Always run
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
