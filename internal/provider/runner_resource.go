package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	datahub "dev.azure.com/AllYourBI/Datahub/_git/go-datahub-sdk.git/pkg/datahub"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var runnerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &runnerResource{}
	_ resource.ResourceWithConfigure   = &runnerResource{}
	_ resource.ResourceWithImportState = &runnerResource{}
)

// NewRunnerResource is a helper function to simplify the provider implementation.
func NewRunnerResource() resource.Resource {
	return &runnerResource{}
}

// runnerResource is the resource implementation.
type runnerResource struct {
	client *datahub.DatahubClient
}

// runnerResourceModel is the resource data model.
type runnerResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	DisplayName       types.String `tfsdk:"display_name"`
	IsGlobal          types.Bool   `tfsdk:"is_global"`
	ClientID          types.String `tfsdk:"client_id"`
	Status            types.String `tfsdk:"status"`
	LastHealthPing    types.String `tfsdk:"last_health_ping"`
	RegisteredAt      types.String `tfsdk:"registered_at"`
	CertificateExpiry types.String `tfsdk:"certificate_expiry"`
	Fingerprint       types.String `tfsdk:"fingerprint"`
	BootstrapToken    types.String `tfsdk:"bootstrap_token"`
}

// Metadata returns the resource type name.
func (r *runnerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner"
}

// Schema defines the schema for the resource.
func (r *runnerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Datahub runner.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the runner.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the runner. Must be unique for the current client.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					runnerNameStringValidator{},
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Display name of the runner.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_global": schema.BoolAttribute{
				Description: "Whether the runner is global.",
				Computed:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "Identifier of the client that owns the runner.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the runner.",
				Computed:    true,
			},
			"last_health_ping": schema.StringAttribute{
				Description: "Timestamp of the last runner health ping.",
				Computed:    true,
			},
			"registered_at": schema.StringAttribute{
				Description: "Timestamp when the runner was registered.",
				Computed:    true,
			},
			"certificate_expiry": schema.StringAttribute{
				Description: "Timestamp when the runner certificate expires.",
				Computed:    true,
			},
			"fingerprint": schema.StringAttribute{
				Description: "Certificate fingerprint for the runner.",
				Computed:    true,
			},
			"bootstrap_token": schema.StringAttribute{
				Description: "Bootstrap token returned when the runner is created.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *runnerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan runnerResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := datahub.CreateRunnerRequest{
		Name: plan.Name.ValueString(),
	}

	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		displayName := plan.DisplayName.ValueString()
		createReq.DisplayName = &displayName
	}

	createdRunner, err := r.client.Runner.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Datahub Runner",
			"Could not create Datahub runner, unexpected error: "+err.Error(),
		)
		return
	}

	state := runnerStateFromCreateResponse(createdRunner)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *runnerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state runnerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idOrName := state.ID.ValueString()
	if parsedID, err := uuid.Parse(idOrName); err == nil {
		idOrName = parsedID.String()
	}

	runner, err := r.client.Runner.Get(ctx, idOrName)
	if err != nil {
		var notFoundErr *datahub.NotFoundError
		if errors.As(err, &notFoundErr) {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Datahub Runner",
			"Could not read Datahub runner ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state = runnerStateFromDetail(runner, normalizeBootstrapToken(state.BootstrapToken))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *runnerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error Updating Datahub Runner",
		"Runner attributes are immutable and should be replaced instead of updated. This indicates an unexpected provider behavior.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *runnerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state runnerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	idOrName := state.ID.ValueString()
	if parsedID, err := uuid.Parse(idOrName); err == nil {
		idOrName = parsedID.String()
	}

	err := r.client.Runner.Delete(ctx, idOrName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Datahub Runner",
			"Could not delete Datahub runner, unexpected error: "+err.Error(),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *runnerResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*datahub.DatahubClient)
}

// ImportState imports the resource state.
func (r *runnerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func runnerStateFromCreateResponse(createdRunner *datahub.CreateRunnerResponse) runnerResourceModel {
	state := runnerStateFromDetail(&createdRunner.RunnerDetail, types.StringValue(createdRunner.BootstrapToken))
	state.BootstrapToken = types.StringValue(createdRunner.BootstrapToken)

	return state
}

func runnerStateFromDetail(runner *datahub.RunnerDetail, bootstrapToken types.String) runnerResourceModel {
	return runnerResourceModel{
		ID:                types.StringValue(runner.ID.String()),
		Name:              types.StringValue(runner.Name),
		DisplayName:       stringValueFromStringPointer(runner.DisplayName),
		IsGlobal:          types.BoolValue(runner.IsGlobal),
		ClientID:          stringValueFromUUIDPointer(runner.ClientID),
		Status:            types.StringValue(runner.Status),
		LastHealthPing:    stringValueFromTimePointer(runner.LastHealthPing),
		RegisteredAt:      types.StringValue(runner.RegisteredAt.Format(time.RFC3339)),
		CertificateExpiry: stringValueFromTimePointer(runner.CertificateExpiry),
		Fingerprint:       types.StringValue(runner.Fingerprint),
		BootstrapToken:    bootstrapToken,
	}
}

func stringValueFromStringPointer(value *string) types.String {
	if value == nil {
		return types.StringNull()
	}

	return types.StringValue(*value)
}

func stringValueFromUUIDPointer(value *uuid.UUID) types.String {
	if value == nil {
		return types.StringNull()
	}

	return types.StringValue(value.String())
}

func stringValueFromTimePointer(value *time.Time) types.String {
	if value == nil {
		return types.StringNull()
	}

	return types.StringValue(value.Format(time.RFC3339))
}

func normalizeBootstrapToken(value types.String) types.String {
	if value.IsNull() || value.IsUnknown() {
		return types.StringValue("")
	}

	return value
}

type runnerNameStringValidator struct{}

func (v runnerNameStringValidator) Description(_ context.Context) string {
	return "value must start with an alphanumeric character and contain only alphanumeric characters, underscores, or hyphens"
}

func (v runnerNameStringValidator) MarkdownDescription(_ context.Context) string {
	return "value must start with an alphanumeric character and contain only alphanumeric characters, underscores, or hyphens"
}

func (v runnerNameStringValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if runnerNameRegex.MatchString(req.ConfigValue.ValueString()) {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid runner name",
		fmt.Sprintf("Runner name %q must match the pattern %q.", req.ConfigValue.ValueString(), runnerNameRegex.String()),
	)
}
