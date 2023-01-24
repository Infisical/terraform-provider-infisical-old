package infisical

import (
	"context"
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
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &infisicalProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &infisicalProvider{}
}

// infisicalProvider is the provider implementation.
type infisicalProvider struct{}

// infisicalProviderModel maps provider schema data to a Go type.
type infisicalProviderModel struct {
	Host   types.String `tfsdk:"host"`
	ApiKey types.String `tfsdk:"api_key"`
}

// Metadata returns the provider type name.
func (p *infisicalProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "infisical"
}

// Schema defines the provider-level schema for configuration data.
func (p *infisicalProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with infisical.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "URI for infisical API. May also be provided via INFISICAL_HOST environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "Username for infisical API. May also be provided via INFISICAL_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a Infisical API client for data sources and resources.
func (p *infisicalProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Infisical client")

	// Retrieve provider data from configuration
	var config infisicalProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.
	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Unknown Infisical API Key",
			"The provider cannot create the Infisical API client as there is an unknown configuration value for the Infisical API Key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the INFISICAL_API_KEY environment variable.",
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
	apiKey := os.Getenv("API_KEY")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.ApiKey.IsNull() {
		apiKey = config.ApiKey.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_key"),
			"Missing Infisical API Key",
			"The provider cannot create the Infisical API client as there is a missing or empty value for the Infisical API Key. "+
				"Set the api_key value in the configuration or use the INFISICAL_API_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "infisical_host", host)
	ctx = tflog.SetField(ctx, "infisical_api_key", apiKey)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "infisical_api_key")

	tflog.Debug(ctx, "Creating Infisical client")

	// Create a new Infisical client using the configuration values
	apiKeyProvider, apiKeyProviderErr := securityprovider.NewSecurityProviderApiKey("header", "X-API-Key", apiKey)
	if apiKeyProviderErr != nil {
		panic(apiKeyProviderErr)
	}

	client, err := ic.NewClient(host, ic.WithRequestEditorFn(apiKeyProvider.Intercept))
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
func (p *infisicalProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationsDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *infisicalProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}
