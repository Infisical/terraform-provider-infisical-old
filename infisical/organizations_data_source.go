package infisical

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	ic "github.com/asheliahut/terraform-provider-infisical/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &organizationsDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationsDataSource{}
)

// NewOrganizationsDataSource is a helper function to simplify the provider implementation.
func NewOrganizationsDataSource() datasource.DataSource {
	return &organizationsDataSource{}
}

// organizationsDataSource is the data source implementation.
type organizationsDataSource struct {
	client *ic.Client
}

// organizationsDataSourceModel maps the data source schema data.
type organizationsDataSourceModel struct {
	Organizations []organizationsModel `tfsdk:"organizations"`
	ID            types.String         `tfsdk:"id"`
}

// organizationsModel maps organizations schema data.
type organizationsModel struct {
	ID          types.String              `tfsdk:"id"`
	Name        types.String              `tfsdk:"name"`
	CreatedAt   types.String              `tfsdk:"created_at"`
	UpdatedAt   types.String              `tfsdk:"updated_at"`
	V           types.Int64               `tfsdk:"v"`
}

// Metadata returns the data source type name.
func (d *organizationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizations"
}

// Schema defines the schema for the data source.
func (d *organizationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of organizations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Placeholder identifier attribute.",
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
						"created_at": schema.StringAttribute{
							Description: "Date of creation for the organization.",
							Computed:    true,
						},
						"updated_at": schema.StringAttribute{
							Description: "Last updated date of the organization.",
							Computed:    true,
						},
						"v": schema.Int64Attribute{
							Description: "Version of the organization.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *organizationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*ic.Client)
}

type MyOrganizationsResponse struct {
	Organizations []struct {
		ID        string `json:"_id"`
		Name      string `json:"name"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		V         int    `json:"__v"`
	} `json:"organizations"`
}

// Read refreshes the Terraform state with the latest data.
func (d *organizationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state organizationsDataSourceModel

	res, err := d.client.GetApiV2UsersMeOrganizations(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Infisical Organizations for User",
			err.Error(),
		)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Infisical Organizations for User",
			err.Error(),
		)
		return
	}

	var data MyOrganizationsResponse
	if err := json.Unmarshal(body, &data); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Decode Infisical Organizations for User",
			err.Error(),
		)
		return
	}

	for _, org := range data.Organizations {
		state.Organizations = append(state.Organizations, organizationsModel{
			ID:          types.StringValue(org.ID),
			Name:        types.StringValue(org.Name),
			CreatedAt:   types.StringValue(org.CreatedAt),
			UpdatedAt:   types.StringValue(org.UpdatedAt),
			V:           types.Int64Value(int64(org.V)),
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
