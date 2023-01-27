package datasource

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	_ datasource.DataSource              = &ProjectsDataSource{}
	_ datasource.DataSourceWithConfigure = &ProjectsDataSource{}
)

// NewProjectsDataSource is a helper function to simplify the provider implementation.
func NewProjectsDataSource() datasource.DataSource {
	return &ProjectsDataSource{}
}

// ProjectsDataSource is the data source implementation.
type ProjectsDataSource struct {
	client *ic.Client
}

// ProjectsDataSourceModel maps the data source schema data.
type ProjectsDataSourceModel struct {
	ID             types.String    `tfsdk:"id"`
	OrganizationId types.String    `tfsdk:"organization_id"`
	Projects       []ProjectsModel `tfsdk:"projects"`
}

// ProjectsModel maps projects schema data.
type ProjectsModel struct {
	ID           types.String              `tfsdk:"id"`
	Name         types.String              `tfsdk:"name"`
	Environments []ProjectEnvironmentModel `tfsdk:"environments"`
}

type ProjectEnvironmentModel struct {
	ID   types.String `tfsdk:"id" json:"_id"`
	Name types.String `tfsdk:"name" json:"name"`
	Slug types.String `tfsdk:"slug" json:"slug"`
}

// Metadata returns the data source type name.
func (d *ProjectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

// Schema defines the schema for the data source.
func (d *ProjectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of projects.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Current Unix timestamp for id.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "Identifier of the organization.",
				Required:    true,
			},
			"projects": schema.ListNestedAttribute{
				Description: "List of projects.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Identifier of the project.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the project.",
							Computed:    true,
						},
						"environments": schema.ListNestedAttribute{
							Description: "List of environments.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Description: "Identifier of the environment.",
										Computed:    true,
									},
									"name": schema.StringAttribute{
										Description: "Name of the environment.",
										Computed:    true,
									},
									"slug": schema.StringAttribute{
										Description: "Slug of the environment.",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

type WorkspacesResponse struct {
	Projects []struct {
		ID           string                `json:"_id"`
		Name         string                `json:"name"`
		Organization string                `json:"organization"`
		Environments EnvironmentReturnList `json:"environments"`
	} `json:"workspaces"`
}

type EnvironmentReturnList []EnvironmentReturn

type EnvironmentReturn struct {
	ID   string `json:"_id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// Read refreshes the Terraform state with the latest data.
func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ProjectsDataSourceModel
	// Get the value of the organization_id attribute
	var organizationId types.String
	diags := req.Config.GetAttribute(ctx, path.Root("organization_id"), &organizationId)

	resp.Diagnostics.Append(diags...)

	res, err := d.client.GetApiV2OrganizationsOrganizationIdWorkspaces(ctx, organizationId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Infisical Projects for User",
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

	var data WorkspacesResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil && err != io.EOF {
		resp.Diagnostics.AddError(
			"Unable to Decode Infisical Projects for User",
			err.Error(),
		)
		return
	}

	for _, proj := range data.Projects {
		var environments []ProjectEnvironmentModel

		for _, env := range proj.Environments {
			environments = append(environments, ProjectEnvironmentModel{
				ID:   types.StringValue(env.ID),
				Name: types.StringValue(env.Name),
				Slug: types.StringValue(env.Slug),
			})
		}
		state.Projects = append(state.Projects, ProjectsModel{
			ID:           types.StringValue(proj.ID),
			Name:         types.StringValue(proj.Name),
			Environments: environments,
		})
	}

	state.OrganizationId = organizationId
	state.ID = types.StringValue(strconv.FormatInt(time.Now().Unix(), 10))

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
