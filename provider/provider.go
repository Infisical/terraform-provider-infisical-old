package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	ic "github.com/asheliahut/terraform-provider-infisical/client"
	ds "github.com/asheliahut/terraform-provider-infisical/datasource"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &InfisicalProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &InfisicalProvider{}
}

// InfisicalProvider is the provider implementation.
type InfisicalProvider struct{}

// InfisicalProviderModel maps provider schema data to a Go type.
type InfisicalProviderModel struct {
	Host     types.String `tfsdk:"host"`
	ApiToken types.String `tfsdk:"api_token"`
}

// Metadata returns the provider type name.
func (p *InfisicalProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "infisical"
}

// Schema defines the provider-level schema for configuration data.
func (p *InfisicalProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with infisical.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for infisical API. May also be provided via INFISICAL_HOST environment variable.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "Username for infisical API. May also be provided via INFISICAL_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a Infisical API client for data sources and resources.
func (p *InfisicalProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Infisical client")

	// Retrieve provider data from configuration
	var config InfisicalProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.ApiToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown Infisical API Token",
			"The provider cannot create the Infisical API client as there is an unknown configuration value for the Infisical API Token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the INFISICAL_API_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	// set host from environment variable and if not set, set to default
	host := os.Getenv("INFISICAL_HOST")
	if host == "" {
		host = "https://infisical.com"
	}
	apiToken := os.Getenv("INFISICAL_API_TOKEN")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.ApiToken.IsNull() {
		apiToken = config.ApiToken.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing Infisical API Token",
			"The provider cannot create the Infisical API client as there is a missing or empty value for the Infisical API Token. "+
				"Set the api_token value in the configuration or use the INFISICAL_API_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "infisical_host", host)
	ctx = tflog.SetField(ctx, "infisical_api_token", apiToken)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "infisical_api_token")

	tflog.Debug(ctx, "Creating Infisical client")

	// Create a new Infisical client using the configuration values
	apiTokenProvider, apiTokenProviderErr := securityprovider.NewSecurityProviderApiKey("header", "X-API-Key", apiToken)
	if apiTokenProviderErr != nil {
		panic(apiTokenProviderErr)
	}

	customProvider := func(ctx context.Context, req *http.Request) error {
		// Just log the request header, nothing else.
		req.Header.Add("accept", "application/json")
		return nil
	}

	client, err := ic.NewClient(host, ic.WithRequestEditorFn(apiTokenProvider.Intercept), ic.WithRequestEditorFn(customProvider))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Infisical API Client",
			"An unexpected error occurred when creating the Infisical API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Infisical Client Error: "+err.Error(),
		)
		return
	}

	// Make the Infisical client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Infisical client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *InfisicalProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		ds.NewOrganizationsDataSource,
		ds.NewProjectsDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *InfisicalProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}
