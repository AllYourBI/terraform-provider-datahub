package provider

import (
	"context"
	"fmt"

	"dev.azure.com/AllYourBI/Datahub/_git/go-datahub-sdk.git/pkg/datahub"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	// "github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &jobResource{}
	_ resource.ResourceWithConfigure   = &jobResource{}
	_ resource.ResourceWithImportState = &jobResource{}
)

// NewJobResource is a helper function to simplify the provider implementation.
func NewJobResource() resource.Resource {
	return &jobResource{}
}

// jobResource is the resource implementation.
// jobResource is the resource implementation.
type jobResource struct {
	client *datahub.DatahubClient
}

// Metadata returns the resource type name.
func (r *jobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_job"
}

// Schema defines the schema for the resource.
// Schema defines the schema for the resource.
func (r *jobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an AYBI Datahub Engine job.",
		Attributes: map[string]schema.Attribute{
			"job_id": schema.StringAttribute{
				Description: "Numeric identifier of the job.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Job, needs to be unique for the current client",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Job type: Full or Incremental",
				Required:    true,
			},
			"image": schema.StringAttribute{
				Description: "Docker image for the job",
				Required:    true,
			},
			"environment": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"secrets": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Sensitive:   true,
			},
			"command": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Command to be executed in the container as a list of arguments",
			},
			"oauth": schema.SingleNestedAttribute{
				Description: "The OAuth config for logging into external service",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"application": schema.StringAttribute{
						Required:    true,
						Description: "the application to connect to like 'exact_online'",
					},
					"flow": schema.StringAttribute{
						Required:    true,
						Description: "the oauth flow type like authorization_code",
					},
					"authorization_url": schema.StringAttribute{
						Required:    true,
						Description: "the full url to start the authorization",
					},
					"token_url": schema.StringAttribute{
						Required:    true,
						Description: "the full URL to fetch the token from",
					},
					"scope": schema.StringAttribute{
						Optional:    true,
						Description: "additional scopes to set for the token request",
					},
					"config_prefix": schema.StringAttribute{
						Required:    true,
						Description: "the prefix for the configuration variables returned from the token response like 'EXACT_ONLINE_",
					},
				},
			},
		},
	}
}

type Environment map[string]string
type Secrets map[string]string
type Command []string

// Create creates the resource and sets the initial Terraform state.
// Create a new resource
func (r *jobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	// Retrieve values from plan
	var job jobResourceModel
	diags := req.Plan.Get(ctx, &job)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var environment map[string]string
	diags = job.Environment.ElementsAs(ctx, &environment, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var secrets map[string]string
	diags = job.Secrets.ElementsAs(ctx, &secrets, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var command []string
	diags = job.Command.ElementsAs(ctx, &command, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var oauth *datahub.OAuthConfig
	if job.OAuth != nil {
		oauth = &datahub.OAuthConfig{
			Application:      job.OAuth.Application.ValueString(),
			Flow:             job.OAuth.Flow.ValueString(),
			AuthorizationUrl: job.OAuth.AuthorizationURL.ValueString(),
			TokenUrl:         job.OAuth.TokenURL.ValueString(),
			Scope:            job.OAuth.Scope.ValueString(),
			ConfigPrefix:     job.OAuth.ConfigPrefix.ValueString(),
		}
	}

	jobRequest := datahub.CreateJobRequest{
		Name:        job.Name.ValueString(),
		JobType:     "full",
		Image:       job.Image.ValueString(),
		Environment: &environment,
		Secrets:     &secrets,
		Command:     &command,
		OAuth:       oauth,
	}

	jobResponse, err := r.client.Job.Create(ctx, jobRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating job",
			"Could not create job, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Job config: %v", job))
	tflog.Debug(ctx, fmt.Sprintf("Create Job object: %v", jobRequest))

	job.JobId = types.StringValue(jobResponse.ID.String())

	// Set state to fully populated data
	diags = resp.State.Set(ctx, job)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *jobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state jobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := uuid.Parse(state.JobId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datahub Job",
			"Could not parse job ID "+state.JobId.ValueString()+": "+err.Error(),
		)
		return
	}

	job, err := r.client.Job.Get(ctx, jobID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datahub Job",
			"Could not read Datahub job ID "+state.JobId.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(job.Name)
	state.Image = types.StringValue(job.Image)

	if len(job.Environment) > 0 {
		// To support jobs that create environment and secrets themselves we don't consider them changes
		// This should be enabled by a feature flag in the future so you can use the terraform config to delete keys from the environment and secrets
		newEnv := dropUntracked(state.Environment, job.Environment)

		state.Environment, diags = types.MapValueFrom(ctx, types.StringType, newEnv)

		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(job.Secrets) > 0 {
		// To support jobs that create environment and secrets themselves we don't consider them changes
		// This should be enabled by a feature flag in the future so you can use the terraform config to delete keys from the environment and secrets
		newSecrets := dropUntracked(state.Secrets, job.Secrets)

		state.Secrets, diags = types.MapValueFrom(ctx, types.StringType, newSecrets)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(job.Command) > 0 {
		state.Command, diags = types.ListValueFrom(ctx, types.StringType, job.Command)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if job.OAuth != nil && job.OAuth.Application != "" && job.OAuth.Flow != "" && job.OAuth.TokenUrl != "" && job.OAuth.AuthorizationUrl != "" && job.OAuth.ConfigPrefix != ""{
		state.OAuth = &jobResourceOauthModel{
			Application:      types.StringValue(job.OAuth.Application),
			Flow:             types.StringValue(job.OAuth.Flow),
			AuthorizationURL: types.StringValue(job.OAuth.AuthorizationUrl),
			TokenURL:         types.StringValue(job.OAuth.TokenUrl),
			ConfigPrefix:     types.StringValue(job.OAuth.ConfigPrefix),
		}
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *jobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan jobResourceModel
	var state jobResourceModel
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

	updateReq := datahub.UpdateOptions{}
	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.Type.Equal(state.Type) {
		typeVal := plan.Type.ValueString()
		updateReq.JobType = &typeVal
	}

	if !plan.Image.Equal(state.Image) {
		image := plan.Image.ValueString()
		updateReq.Image = &image
	}

	if !plan.Environment.Equal(state.Environment) {
		var environment map[string]string
		diags = plan.Environment.ElementsAs(ctx, &environment, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.Environment = &environment

	}

	if !plan.Secrets.Equal(state.Secrets) {
		var secrets map[string]string
		diags = plan.Secrets.ElementsAs(ctx, &secrets, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.Secrets = &secrets
	}

	if !plan.Command.Equal(state.Command) {
		var command []string
		diags = plan.Command.ElementsAs(ctx, &command, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.Command = &command
	}

	if (state.OAuth == nil && plan.OAuth != nil) || (plan.OAuth != nil && (!plan.OAuth.Application.Equal(state.OAuth.Application) ||
		!plan.OAuth.Flow.Equal(state.OAuth.Flow) ||
		!plan.OAuth.TokenURL.Equal(state.OAuth.TokenURL) ||
		!plan.OAuth.AuthorizationURL.Equal(state.OAuth.AuthorizationURL) ||
		!plan.OAuth.Scope.Equal(state.OAuth.Scope) ||
		!plan.OAuth.ConfigPrefix.Equal(state.OAuth.ConfigPrefix))) {

		updateReq.OAuth = &datahub.OAuthConfig{
			Application:      plan.OAuth.Application.ValueString(),
			Flow:             plan.OAuth.Flow.ValueString(),
			TokenUrl:         plan.OAuth.TokenURL.ValueString(),
			AuthorizationUrl: plan.OAuth.AuthorizationURL.ValueString(),
			Scope:            plan.OAuth.Scope.ValueString(),
			ConfigPrefix:     plan.OAuth.ConfigPrefix.ValueString(),
		}
	}

	jobID, err := uuid.Parse(state.JobId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Datahub Job",
			"Could not parse job ID "+state.JobId.ValueString()+": "+err.Error(),
		)
		return
	}

	_, err = r.client.Job.Update(ctx, jobID, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Datahub Job",
			"unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *jobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state jobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobID, err := uuid.Parse(state.JobId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Datahub Job",
			"Could not parse job ID "+state.JobId.ValueString()+": "+err.Error(),
		)
		return
	}

	// Delete existing job
	err = r.client.Job.Delete(ctx, jobID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Datahub Job",
			"Could not delete job, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *jobResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*datahub.DatahubClient)
}

func (r *jobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("job_id"), req, resp)
}

type jobResourceModel struct {
	JobId       types.String           `tfsdk:"job_id"`
	Name        types.String           `tfsdk:"name"`
	Type        types.String           `tfsdk:"type"`
	Image       types.String           `tfsdk:"image"`
	Environment types.Map              `tfsdk:"environment"`
	Secrets     types.Map              `tfsdk:"secrets"`
	Command     types.List             `tfsdk:"command"`
	OAuth       *jobResourceOauthModel `tfsdk:"oauth"`
}

type jobResourceOauthModel struct {
	Application      types.String `tfsdk:"application"`
	Flow             types.String `tfsdk:"flow"`
	AuthorizationURL types.String `tfsdk:"authorization_url"`
	TokenURL         types.String `tfsdk:"token_url"`
	Scope            types.String `tfsdk:"scope"`
	ConfigPrefix     types.String `tfsdk:"config_prefix"`
}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func dropUntracked(state types.Map, new map[string]string) map[string]string {
	originalKeys := []string{}
	for key, _ := range state.Elements() {
		originalKeys = append(originalKeys, key)
	}

	newKeys := []string{}
	for key, _ := range new {
		newKeys = append(newKeys, key)
	}

	toDel := difference(newKeys, originalKeys)
	for _, keyToDel := range toDel {
		delete(new, keyToDel)
	}

	return new
}
