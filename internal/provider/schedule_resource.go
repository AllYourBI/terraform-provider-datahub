package provider

import (
	"context"
	"fmt"

	"dev.azure.com/AllYourBI/Datahub/_git/go-datahub-sdk.git/pkg/datahub"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &scheduleResource{}
	_ resource.ResourceWithConfigure   = &scheduleResource{}
	_ resource.ResourceWithImportState = &scheduleResource{}
)

// NewScheduleResource is a helper function to simplify the provider implementation.
func NewScheduleResource() resource.Resource {
	return &scheduleResource{}
}

// scheduleResource is the resource implementation.
type scheduleResource struct {
	client *datahub.DatahubClient
}

// Metadata returns the resource type name.
func (r *scheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

// Schema defines the schema for the resource.
func (r *scheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Datahub schedule.",
		Attributes: map[string]schema.Attribute{
			"schedule_id": schema.StringAttribute{
				Description: "Identifier of the schedule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule": schema.StringAttribute{
				Description: "Cron expression for the schedule.",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the schedule is enabled. Default is true.",
				Optional:    true,
				Computed:    true,
			},
			"maximum_drift": schema.Int64Attribute{
				Description: "Maximum drift in seconds. Default is 0.",
				Optional:    true,
				Computed:    true,
			},
			"backfill": schema.BoolAttribute{
				Description: "Whether to backfill missed runs. Default is false.",
				Optional:    true,
				Computed:    true,
			},
			"bindings": schema.ListNestedAttribute{
				Description: "List of job bindings for this schedule.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"job_id": schema.StringAttribute{
							Description: "ID of the job to bind to this schedule.",
							Required:    true,
						},
						"environment": schema.MapAttribute{
							Description: "Environment variables to override for this job binding.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"secrets": schema.MapAttribute{
							Description: "Secrets to override for this job binding.",
							Optional:    true,
							Sensitive:   true,
							ElementType: types.StringType,
						},
						"command": schema.ListAttribute{
							Description: "Command to override for this job binding.",
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"start_at": schema.StringAttribute{
				Description: "Start date/time of the schedule in RFC3339 format.",
				Optional:    true,
			},
			"end_at": schema.StringAttribute{
				Description: "End date/time of the schedule in RFC3339 format.",
				Optional:    true,
			},
		},
	}
}

// scheduleResourceModel is the resource data model.
type scheduleResourceModel struct {
	ScheduleID   types.String      `tfsdk:"schedule_id"`
	Schedule     types.String      `tfsdk:"schedule"`
	Enabled      types.Bool        `tfsdk:"enabled"`
	MaximumDrift types.Int64       `tfsdk:"maximum_drift"`
	Backfill     types.Bool        `tfsdk:"backfill"`
	Bindings     []scheduleBinding `tfsdk:"bindings"`
	StartAt      types.String      `tfsdk:"start_at"`
	EndAt        types.String      `tfsdk:"end_at"`
}

type scheduleBinding struct {
	JobID       types.String `tfsdk:"job_id"`
	Environment types.Map    `tfsdk:"environment"`
	Secrets     types.Map    `tfsdk:"secrets"`
	Command     types.List   `tfsdk:"command"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *scheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan scheduleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build SDK bindings slice
	sdkBindings := []datahub.ScheduleBinding{}
	for _, binding := range plan.Bindings {
		jobID, err := uuid.Parse(binding.JobID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Creating Schedule",
				fmt.Sprintf("Could not parse job ID %s: %s", binding.JobID.ValueString(), err),
			)
			return
		}

		sdkBinding := datahub.ScheduleBinding{
			JobID: jobID,
		}

		// Handle environment if set
		if !binding.Environment.IsNull() && !binding.Environment.IsUnknown() {
			var environment map[string]string
			diags = binding.Environment.ElementsAs(ctx, &environment, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			sdkBinding.Environment = &environment
		}

		// Handle secrets if set
		if !binding.Secrets.IsNull() && !binding.Secrets.IsUnknown() {
			var secrets map[string]string
			diags = binding.Secrets.ElementsAs(ctx, &secrets, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			sdkBinding.Secrets = &secrets
		}

		// Handle command if set
		if !binding.Command.IsNull() && !binding.Command.IsUnknown() {
			var command []string
			diags = binding.Command.ElementsAs(ctx, &command, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			sdkBinding.Command = &command
		}

		sdkBindings = append(sdkBindings, sdkBinding)
	}

	// Build the request
	scheduleRequest := datahub.ScheduleRequest{
		Schedule: plan.Schedule.ValueString(),
		Bindings: sdkBindings,
	}

	// Set enabled if provided, otherwise SDK will apply default
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		enabled := plan.Enabled.ValueBool()
		scheduleRequest.Enabled = &enabled
	}

	// Set maximum_drift if provided, otherwise SDK will apply default
	if !plan.MaximumDrift.IsNull() && !plan.MaximumDrift.IsUnknown() {
		maximumDrift := int(plan.MaximumDrift.ValueInt64())
		scheduleRequest.MaximumDrift = &maximumDrift
	}

	// Set backfill if provided, otherwise SDK will apply default
	if !plan.Backfill.IsNull() && !plan.Backfill.IsUnknown() {
		backfill := plan.Backfill.ValueBool()
		scheduleRequest.Backfill = &backfill
	}

	// Parse and set start_at if provided
	// TODO: Add time parsing for start_at and end_at

	tflog.Debug(ctx, fmt.Sprintf("Creating schedule with request: %v", scheduleRequest))

	// Call API to create the schedule
	schedule, err := r.client.Schedule.Create(ctx, scheduleRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Schedule",
			"Could not create schedule, unexpected error: "+err.Error(),
		)
		return
	}

	// Map the response to the plan
	plan.ScheduleID = types.StringValue(schedule.ScheduleID.String())
	plan.Schedule = types.StringValue(schedule.Schedule)
	plan.Enabled = types.BoolValue(schedule.Enabled)

	if schedule.MaximumDrift != nil {
		plan.MaximumDrift = types.Int64Value(int64(*schedule.MaximumDrift))
	} else {
		plan.MaximumDrift = types.Int64Value(0)
	}

	plan.Backfill = types.BoolValue(schedule.Backfill)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *scheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state scheduleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID, err := uuid.Parse(state.ScheduleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datahub Schedule",
			"Could not parse schedule ID "+state.ScheduleID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Call API to get the schedule
	schedule, err := r.client.Schedule.Get(ctx, scheduleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datahub Schedule",
			"Could not read Datahub schedule ID "+state.ScheduleID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with the fresh data from API
	state.Schedule = types.StringValue(schedule.Schedule)
	state.Enabled = types.BoolValue(schedule.Enabled)

	if schedule.MaximumDrift != nil {
		state.MaximumDrift = types.Int64Value(int64(*schedule.MaximumDrift))
	} else {
		state.MaximumDrift = types.Int64Value(0)
	}

	state.Backfill = types.BoolValue(schedule.Backfill)

	// TODO: Handle bindings update from API

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *scheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan, state scheduleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID, err := uuid.Parse(state.ScheduleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Datahub Schedule",
			"Could not parse schedule ID "+state.ScheduleID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Build SDK bindings slice
	sdkBindings := []datahub.ScheduleBinding{}
	for _, binding := range plan.Bindings {
		jobID, err := uuid.Parse(binding.JobID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Schedule",
				fmt.Sprintf("Could not parse job ID %s: %s", binding.JobID.ValueString(), err),
			)
			return
		}

		sdkBinding := datahub.ScheduleBinding{
			JobID: jobID,
		}

		// Handle environment if set
		if !binding.Environment.IsNull() && !binding.Environment.IsUnknown() {
			var environment map[string]string
			diags = binding.Environment.ElementsAs(ctx, &environment, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			sdkBinding.Environment = &environment
		}

		// Handle secrets if set
		if !binding.Secrets.IsNull() && !binding.Secrets.IsUnknown() {
			var secrets map[string]string
			diags = binding.Secrets.ElementsAs(ctx, &secrets, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			sdkBinding.Secrets = &secrets
		}

		// Handle command if set
		if !binding.Command.IsNull() && !binding.Command.IsUnknown() {
			var command []string
			diags = binding.Command.ElementsAs(ctx, &command, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			sdkBinding.Command = &command
		}

		sdkBindings = append(sdkBindings, sdkBinding)
	}

	// Build the update request
	scheduleRequest := datahub.ScheduleRequest{
		Schedule: plan.Schedule.ValueString(),
		Bindings: sdkBindings,
	}

	// Set enabled if provided
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		enabled := plan.Enabled.ValueBool()
		scheduleRequest.Enabled = &enabled
	}

	// Set maximum_drift if provided
	if !plan.MaximumDrift.IsNull() && !plan.MaximumDrift.IsUnknown() {
		maximumDrift := int(plan.MaximumDrift.ValueInt64())
		scheduleRequest.MaximumDrift = &maximumDrift
	}

	// Set backfill if provided
	if !plan.Backfill.IsNull() && !plan.Backfill.IsUnknown() {
		backfill := plan.Backfill.ValueBool()
		scheduleRequest.Backfill = &backfill
	}

	// Parse and set start_at if provided
	// TODO: Add time parsing for start_at and end_at

	// Call API to update the schedule
	schedule, err := r.client.Schedule.Update(ctx, scheduleID, scheduleRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Datahub Schedule",
			"Could not update Datahub schedule, unexpected error: "+err.Error(),
		)
		return
	}

	// Map the response to the plan
	plan.ScheduleID = types.StringValue(schedule.ScheduleID.String())
	plan.Schedule = types.StringValue(schedule.Schedule)
	plan.Enabled = types.BoolValue(schedule.Enabled)

	if schedule.MaximumDrift != nil {
		plan.MaximumDrift = types.Int64Value(int64(*schedule.MaximumDrift))
	} else {
		plan.MaximumDrift = types.Int64Value(0)
	}

	plan.Backfill = types.BoolValue(schedule.Backfill)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *scheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state scheduleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scheduleID, err := uuid.Parse(state.ScheduleID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Datahub Schedule",
			"Could not parse schedule ID "+state.ScheduleID.ValueString()+": "+err.Error(),
		)
		return
	}

	err = r.client.Schedule.Delete(ctx, scheduleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Datahub Schedule",
			"Could not delete schedule, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *scheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*datahub.DatahubClient)
}

func (r *scheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("schedule_id"), req, resp)
}
