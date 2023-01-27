package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ic "github.com/asheliahut/terraform-provider-infisical/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &OrganizationsDataSource{}
	_ datasource.DataSourceWithConfigure = &OrganizationsDataSource{}
)

// NewOrganizationsDataSource is a helper function to simplify the provider implementation.
func NewOrganizationsDataSource() datasource.DataSource {
	return &OrganizationsDataSource{}
}

// OrganizationsDataSource is the data source implementation.
type OrganizationsDataSource struct {
	client *ic.Client
}

// OrganizationsDataSourceModel maps the data source schema data.
type OrganizationsDataSourceModel struct {
	Organizations []OrganizationsModel `tfsdk:"organizations"`
	ID            types.String         `tfsdk:"id"`
}

// OrganizationsModel maps organizations schema data.
type OrganizationsModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *OrganizationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizations"
}

// Schema defines the schema for the data source.
func (d *OrganizationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of organizations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Current Unix timestamp for id.",
				Computed:    true,
			},
			"organizations": schema.ListNestedAttribute{
				Description: "List of organizations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Identifier of the organization.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the organization.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *OrganizationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ic.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

type MyOrganizationsResponse struct {
	Organizations []struct {
		ID   string `json:"_id"`
		Name string `json:"name"`
	} `json:"organizations"`
}

// Read refreshes the Terraform state with the latest data.
func (d *OrganizationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OrganizationsDataSourceModel

	res, err := d.client.GetApiV2UsersMeOrganizations(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Infisical Organizations for User",
			err.Error(),
		)
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to close body reader for User",
				err.Error(),
			)
			return
		}
	}(res.Body)

	var data MyOrganizationsResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil && err != io.EOF {
		resp.Diagnostics.AddError(
			"Unable to Decode Infisical Organizations for User",
			err.Error(),
		)
		return
	}

	for _, org := range data.Organizations {
		state.Organizations = append(state.Organizations, OrganizationsModel{
			ID:   types.StringValue(org.ID),
			Name: types.StringValue(org.Name),
		})
	}

	state.ID = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
