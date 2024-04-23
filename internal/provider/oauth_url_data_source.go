package provider

import (
	"context"
	"dev.azure.com/AllYourBI/Datahub/_git/go-datahub-sdk.git/pkg/datahub"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &oauthDataSource{}
	_ datasource.DataSourceWithConfigure = &oauthDataSource{}
)

// NewoauthDataSource is a helper function to simplify the provider implementation.
func NewOAuthURLDataSource() datasource.DataSource {
	return &oauthDataSource{}

}

type oauthDataSource struct {
	client *datahub.DatahubClient
}

func (d *oauthDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_oauth_url"
}

// Schema defines the schema for the data source.
func (d *oauthDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches an OAuth URL for starting the OAuth flow",
		Attributes: map[string]schema.Attribute{
			"job_id": schema.StringAttribute{
				Description: "Placeholder identifier attribute.",
				Required:    true,
			},
			"redirect": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *oauthDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config oauthDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := uuid.Parse(config.JobID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to parse job_id",
			err.Error(),
		)
		return
	}

	oauthResponse, err := d.client.Job.GetOAuthRedirect(ctx, jobID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Get OAuth redirect url for this job",
			err.Error(),
		)
		return
	}

	config.Redirect = types.StringValue(oauthResponse.Redirect)

	// Set state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *oauthDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*datahub.DatahubClient)
}

// oauthDataSourceModel maps the data source schema data.
type oauthDataSourceModel struct {
	JobID    types.String `tfsdk:"job_id"`
	Redirect types.String `tfsdk:"redirect"`
}
