package provider

import (
	"context"
	"time"

	datahub "dev.azure.com/AllYourBI/Datahub/_git/go-datahub-sdk.git/pkg/datahub"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &runnersDataSource{}
	_ datasource.DataSourceWithConfigure = &runnersDataSource{}
)

// NewRunnersDataSource is a helper function to simplify the provider implementation.
func NewRunnersDataSource() datasource.DataSource {
	return &runnersDataSource{}
}

type runnersDataSource struct {
	client *datahub.DatahubClient
}

func (d *runnersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runners"
}

// Schema defines the schema for the data source.
func (d *runnersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches Datahub runners.",
		Attributes: map[string]schema.Attribute{
			"runners": schema.ListNestedAttribute{
				Description: "List of runners returned by the Datahub API.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"display_name": schema.StringAttribute{
							Computed: true,
						},
						"is_global": schema.BoolAttribute{
							Computed: true,
						},
						"client_id": schema.StringAttribute{
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
						"last_health_ping": schema.StringAttribute{
							Computed: true,
						},
						"registered_at": schema.StringAttribute{
							Computed: true,
						},
						"certificate_expiry": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *runnersDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	runnersResponse, err := d.client.Runner.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to list Datahub runners",
			err.Error(),
		)
		return
	}

	state := runnersDataSourceModel{
		Runners: make([]runnerModel, 0),
	}

	if runnersResponse != nil {
		state.Runners = make([]runnerModel, 0, len(runnersResponse.Runners))
		for _, runner := range runnersResponse.Runners {
			state.Runners = append(state.Runners, runnerModel{
				ID:                types.StringValue(runner.ID.String()),
				Name:              types.StringValue(runner.Name),
				DisplayName:       stringValueFromStringPointer(runner.DisplayName),
				IsGlobal:          types.BoolValue(runner.IsGlobal),
				ClientID:          stringValueFromUUIDPointer(runner.ClientID),
				Status:            types.StringValue(runner.Status),
				LastHealthPing:    stringValueFromTimePointer(runner.LastHealthPing),
				RegisteredAt:      types.StringValue(runner.RegisteredAt.Format(time.RFC3339)),
				CertificateExpiry: stringValueFromTimePointer(runner.CertificateExpiry),
			})
		}
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Configure adds the provider configured client to the data source.
func (d *runnersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*datahub.DatahubClient)
}

type runnersDataSourceModel struct {
	Runners []runnerModel `tfsdk:"runners"`
}

type runnerModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	DisplayName       types.String `tfsdk:"display_name"`
	IsGlobal          types.Bool   `tfsdk:"is_global"`
	ClientID          types.String `tfsdk:"client_id"`
	Status            types.String `tfsdk:"status"`
	LastHealthPing    types.String `tfsdk:"last_health_ping"`
	RegisteredAt      types.String `tfsdk:"registered_at"`
	CertificateExpiry types.String `tfsdk:"certificate_expiry"`
}
