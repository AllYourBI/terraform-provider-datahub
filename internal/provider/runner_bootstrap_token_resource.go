package provider

import (
	"context"

	datahub "dev.azure.com/AllYourBI/Datahub/_git/go-datahub-sdk.git/pkg/datahub"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &runnerBootstrapTokenResource{}
	_ resource.ResourceWithConfigure = &runnerBootstrapTokenResource{}
)

// NewRunnerBootstrapTokenResource is a helper function to simplify the provider implementation.
func NewRunnerBootstrapTokenResource() resource.Resource {
	return &runnerBootstrapTokenResource{}
}

// runnerBootstrapTokenResource is the resource implementation.
type runnerBootstrapTokenResource struct {
	client *datahub.DatahubClient
}

// runnerBootstrapTokenResourceModel is the resource data model.
type runnerBootstrapTokenResourceModel struct {
	ID             types.String `tfsdk:"id"`
	RunnerID       types.String `tfsdk:"runner_id"`
	Name           types.String `tfsdk:"name"`
	BootstrapToken types.String `tfsdk:"bootstrap_token"`
	Triggers       types.Map    `tfsdk:"triggers"`
}

// Metadata returns the resource type name.
func (r *runnerBootstrapTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner_bootstrap_token"
}

// Schema defines the schema for the resource.
func (r *runnerBootstrapTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Rotates the bootstrap token for a Datahub runner. Replacing this resource invalidates existing runner certificates and requires the runner to bootstrap again.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the runner whose bootstrap token was regenerated.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"runner_id": schema.StringAttribute{
				Description: "UUID or name of the runner to reset the bootstrap token for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the runner.",
				Computed:    true,
			},
			"bootstrap_token": schema.StringAttribute{
				Description: "New bootstrap token returned after regeneration.",
				Computed:    true,
				Sensitive:   true,
			},
			"triggers": schema.MapAttribute{
				Description: "Arbitrary map of values that, when changed, force bootstrap token regeneration.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Create regenerates the runner bootstrap token and sets the Terraform state.
func (r *runnerBootstrapTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state runnerBootstrapTokenResourceModel
	diags := req.Plan.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	regeneratedRunner, err := r.client.Runner.RegenerateBootstrapToken(ctx, state.RunnerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Regenerating Datahub Runner Bootstrap Token",
			"Could not regenerate Datahub runner bootstrap token, unexpected error: "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(regeneratedRunner.ID.String())
	state.Name = types.StringValue(regeneratedRunner.Name)
	state.BootstrapToken = types.StringValue(regeneratedRunner.BootstrapToken)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Read keeps the existing Terraform state because the bootstrap token cannot be re-read from the API.
func (r *runnerBootstrapTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state runnerBootstrapTokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update is not supported because all configurable attributes force replacement.
func (r *runnerBootstrapTokenResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error Updating Datahub Runner Bootstrap Token",
		"Runner bootstrap tokens are rotated by replacing the resource. This indicates an unexpected provider behavior.",
	)
}

// Delete is a no-op because the token lifecycle is tied to the runner itself.
func (r *runnerBootstrapTokenResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// Configure adds the provider configured client to the resource.
func (r *runnerBootstrapTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*datahub.DatahubClient)
}
