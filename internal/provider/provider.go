package provider

import (
	"context"
	"terraform-provider-datahub/internal/datahub"

	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &datahubProvider{}
)

type datahubProviderModel struct {
	BaseURL      types.String `tfsdk:"base_url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

// New is a helper function to simplify provider server and testing implementation.
func New() provider.Provider {
	return &datahubProvider{}
}

// DatahubProvider is the provider implementation.
type datahubProvider struct{}

// Metadata returns the provider type name.
func (p *datahubProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "datahub"
}

// Schema defines the provider-level schema for configuration data.
func (p *datahubProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Datahub.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "Base URL for the datahub api, like: https://api.datahub.allyourbi.nl",
				Optional:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "Client ID for the Datahub Engine API. May also be provided via DATAHUB_CLIENT_ID environment variable.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "Client Secret for the Datahub Engine API. May also be provided via DATAHUB_CLIENT_SECRET environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure prepares a Datahub API client for data sources and resources.
func (p *datahubProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Datahub client")

	// Retrieve provider data from configuration
	var config datahubProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.BaseURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("BaseURL"),
			"Unknown Datahub API BaseURL",
			"The provider cannot create the Datahub API client as there is an unknown configuration value for the Datahub API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DATAHUB_HOST environment variable.",
		)
	}

	if config.ClientID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown Datahub API ClientID",
			"The provider cannot create the Datahub API client as there is an unknown configuration value for the Datahub API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DATAHUB_USERNAME environment variable.",
		)
	}

	if config.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Unknown Datahub API ClientSecret",
			"The provider cannot create the Datahub API client as there is an unknown configuration value for the Datahub API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the DATAHUB_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	base_url := os.Getenv("DATAHUB_HOST")
	client_id := os.Getenv("DATAHUB_CLIENT_ID")
	client_secret := os.Getenv("DATAHUB_CLIENT_SECRET")

	if !config.BaseURL.IsNull() {
		base_url = config.BaseURL.ValueString()
	}

	if !config.ClientID.IsNull() {
		client_id = config.ClientID.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		client_secret = config.ClientSecret.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if base_url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Datahub API Host",
			"The provider cannot create the Datahub API client as there is a missing or empty value for the Datahub API host. "+
				"Set the host value in the configuration or use the DATAHUB_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if client_id == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Datahub API Username",
			"The provider cannot create the Datahub API client as there is a missing or empty value for the Datahub API username. "+
				"Set the username value in the configuration or use the DATAHUB_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if client_secret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Datahub API Password",
			"The provider cannot create the Datahub API client as there is a missing or empty value for the Datahub API password. "+
				"Set the password value in the configuration or use the DATAHUB_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "DATAHUB_BASE_URL", base_url)
	ctx = tflog.SetField(ctx, "DATAHUB_CLIENT_ID", client_id)
	ctx = tflog.SetField(ctx, "DATAHUB_CLIENT_SECRET", client_secret)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "DATAHUB_CLIENT_SECRET")

	tflog.Debug(ctx, "Creating Datahub client")

	// Create a new Datahub client using the configuration values
	client := datahub.NewDatahubClient(base_url, client_id, client_secret)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Create Datahub API Client",
	// 		"An unexpected error occurred when creating the Datahub API client. "+
	// 			"If the error is not clear, please contact the provider developers.\n\n"+
	// 			"Datahub Client Error: "+err.Error(),
	// 	)
	// 	return
	// }

	// Make the Datahub client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Datahub client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *datahubProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewCoffeesDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *datahubProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewJobResource,
	}
}
