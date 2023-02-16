package provider

import (
	"context"
	// "encoding/json"
	"fmt"

	"terraform-provider-datahub/internal/datahub"

	// "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/exp/slices"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	// "github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &initRunResource{}
	_ resource.ResourceWithConfigure = &initRunResource{}
	// _ resource.ResourceWithImportState = &initRunResource{}
)

// NewInitRunResource is a helper function to simplify the provider implementation.
func NewInitRunResource() resource.Resource {
	return &initRunResource{}
}

// initRunResource is the resource implementation.
// initRunResource is the resource implementation.
type initRunResource struct {
	client *datahub.DatahubClient
}

// Metadata returns the resource type name.
func (r *initRunResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_init_run"
}

// Schema defines the schema for the resource.
// Schema defines the schema for the resource.
func (r *initRunResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the execution of an initialisation script",
		Attributes: map[string]schema.Attribute{
			"run_id": schema.StringAttribute{
				Description: "Numeric identifier of the init run.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"job_id": schema.StringAttribute{
				Description: "Numeric identifier of the created job",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the InitRun, needs to be unique for the current client",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// "type": schema.StringAttribute{
			// 	Description: "InitRun type: Full or Incremental",
			// 	Required:    true,
			// },
			"image": schema.StringAttribute{
				Description: "Docker image for the job",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"secrets": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"command": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Command to be executed in the container as a list of arguments",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			// "oauth": schema.SingleNestedAttribute{
			// 	Description: "The OAuth config for logging into external service",
			// 	Optional:    true,
			// 	Attributes: map[string]schema.Attribute{
			// 		"application": schema.StringAttribute{
			// 			Required:    true,
			// 			Description: "the application to connect to like 'exact_online'",
			// 		},
			// 		"flow": schema.StringAttribute{
			// 			Required:    true,
			// 			Description: "the oauth flow type like authorization_code",
			// 		},
			// 		"authorization_url": schema.StringAttribute{
			// 			Required:    true,
			// 			Description: "the full url to start the authorization",
			// 		},
			// 		"token_url": schema.StringAttribute{
			// 			Required:    true,
			// 			Description: "the full URL to fetch the token from",
			// 		},
			// 		"scope": schema.StringAttribute{
			// 			Optional:    true,
			// 			Description: "additional scopes to set for the token request",
			// 		},
			// 		"config_prefix": schema.StringAttribute{
			// 			Required:    true,
			// 			Description: "the prefix for the configuration variables returned from the token response like 'EXACT_ONLINE_",
			// 		},
			// 	},
			// },
			"status": schema.StringAttribute{
				Description: "Status of the init run can be either OK or FAILED",
				Computed:    true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
// Create a new resource
func (r *initRunResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var run runResourceModel
	diags := req.Plan.Get(ctx, &run)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var environment Environment
	diags = run.Environment.ElementsAs(ctx, &environment, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var secrets Secrets
	diags = run.Secrets.ElementsAs(ctx, &secrets, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var command Command
	diags = run.Command.ElementsAs(ctx, &command, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobRequest := datahub.CreateJobRequest{
		Name:        run.Name.ValueString(),
		Type:        "full",
		Image:       run.Image.ValueString(),
		Environment: environment,
		Secrets:     secrets,
		Command:     command,
	}

	jobResponse, err := r.client.CreateJob(ctx, jobRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating job",
			"Could not create job, unexpected error: "+err.Error(),
		)
		return
	}

	// tflog.Debug(ctx, fmt.Sprintf("Job config: %v", job))
	tflog.Debug(ctx, fmt.Sprintf("Create Job object: %v", jobRequest))

	run.JobID = types.StringValue(jobResponse.JobID)

	runRequest := datahub.CreateRunRequest{
		JobID: jobResponse.JobID,
	}

	runResponse, err := r.client.CreateRun(ctx, runRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating run",
			"Could not create run, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("InitRun config: %v", run))
	tflog.Debug(ctx, fmt.Sprintf("Create InitRun object: %v", runRequest))

	run.RunID = types.StringValue(runResponse.RunID)
	run.Status = types.StringValue(InitRunStatusOK)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, run)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *initRunResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state runResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// 	// Get refreshed job value from HashiCups
	runStatus, err := r.client.GetRunStatus(ctx, state.RunID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datahub InitRun Status",
			"Could not read Datahub job ID "+state.RunID.ValueString()+": "+err.Error(),
		)
		return
	}

	if slices.Contains([]string{"failed", "cancelled", "rejected"}, runStatus.Status) {
		// state.Status = types.StringValue(InitRunStatusFailed)
		resp.State.RemoveResource(ctx)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *initRunResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Error updating initialise run",
		"An initialise run should be immutable, so it should be recreated and not updated. This is an error in the plugin",
	)
	
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *initRunResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// // Retrieve values from state
	// 	var state runResourceModel
	// 	diags := req.State.Get(ctx, &state)
	// 	resp.Diagnostics.Append(diags...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}

	// 	// Delete existing job
	// 	err := r.client.DeleteInitRun(state.ID.ValueString())
	// 	if err != nil {
	// 		resp.Diagnostics.AddError(
	// 			"Error Deleting HashiCups InitRun",
	// 			"Could not delete job, unexpected error: "+err.Error(),
	// 		)
	// 		return
	// 	}
}

func (r *initRunResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*datahub.DatahubClient)
}

// func (r *initRunResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
// 	// Retrieve import ID and save to id attribute
// 	resource.ImportStatePassthroughID(ctx, path.Root("job_id"), req, resp)
// }

type runResourceModel struct {
	JobID types.String `tfsdk:"job_id"`
	RunID types.String `tfsdk:"run_id"`
	Name  types.String `tfsdk:"name"`
	// Type        types.String           `tfsdk:"type"`
	Image       types.String `tfsdk:"image"`
	Environment types.Map    `tfsdk:"environment"`
	Secrets     types.Map    `tfsdk:"secrets"`
	Command     types.List   `tfsdk:"command"`
	Status      types.String `tfsdk:"status"`
	// OAuth       *runResourceOauthModel `tfsdk:"oauth"`
}

type runResourceOauthModel struct {
	Application      types.String `tfsdk:"application"`
	Flow             types.String `tfsdk:"flow"`
	AuthorizationURL types.String `tfsdk:"authorization_url"`
	TokenURL         types.String `tfsdk:"token_url"`
	Scope            types.String `tfsdk:"scope"`
	ConfigPrefix     types.String `tfsdk:"config_prefix"`
}

const InitRunStatusOK = "OK"
const InitRunStatusFailed = "FAILED"
