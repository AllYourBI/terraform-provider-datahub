package terraform

import (
	"context"
	"encoding/json"
	"fmt"

	// "fmt"

	// "strconv"
	"terraform-provider-datahub/internal/datahub"
	// "time"

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
			// "items": schema.ListNestedAttribute{
			//     Description: "List of items in the job.",
			//     Required:    true,
			//     NestedObject: schema.NestedAttributeObject{
			//         Attributes: map[string]schema.Attribute{
			//             "quantity": schema.Int64Attribute{
			//                 Description: "Count of this item in the job.",
			//                 Required:    true,
			//             },
			//             "coffee": schema.SingleNestedAttribute{
			//                 Description: "Coffee item in the job.",
			//                 Required:    true,
			//                 Attributes: map[string]schema.Attribute{
			//                     "id": schema.Int64Attribute{
			//                         Description: "Numeric identifier of the coffee.",
			//                         Required:    true,
			//                     },
			//                     "name": schema.StringAttribute{
			//                         Description: "Product name of the coffee.",
			//                         Computed:    true,
			//                     },
			//                     "teaser": schema.StringAttribute{
			//                         Description: "Fun tagline for the coffee.",
			//                         Computed:    true,
			//                     },
			//                     "description": schema.StringAttribute{
			//                         Description: "Product description of the coffee.",
			//                         Computed:    true,
			//                     },
			//                     "price": schema.Float64Attribute{
			//                         Description: "Suggested cost of the coffee.",
			//                         Computed:    true,
			//                     },
			//                     "image": schema.StringAttribute{
			//                         Description: "URI for an image of the coffee.",
			//                         Computed:    true,
			//                     },
			//                 },
			//             },
			//         },
			//     },
			// },
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

	var environment Environment
	diags = job.Environment.ElementsAs(ctx, &environment, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	} 

	var secrets Secrets
	diags = job.Secrets.ElementsAs(ctx, &secrets, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	} 

	var command Command
	diags = job.Command.ElementsAs(ctx, &command, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	} 


	jobRequest := datahub.CreateJobRequest{
		Name:        job.Name.ValueString(),
		Type:        "full",
		Image:       job.Image.ValueString(),
		Environment: environment,
		Secrets:     secrets,
		Command:     command,
	}

	jobJSON, err := json.Marshal(jobRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error marshalling to JSON",
			"Could not create job, unexpected error: "+err.Error(),
		)
		return
	}
	tflog.Debug(ctx, string(jobJSON))

	jobResponse, err := r.client.CreateJob(ctx, jobRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating job",
			"Could not create job, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Job config: %v", job))
	tflog.Debug(ctx, fmt.Sprintf("Create Job object: %v", jobRequest))

	job.JobId = types.StringValue(jobResponse.JobID)

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

	// 	// Get refreshed job value from HashiCups
	job, err := r.client.GetJob(ctx, state.JobId.ValueString())
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
		state.Environment, diags = types.MapValueFrom(ctx, types.StringType, job.Environment)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	
	if len(job.Secrets) > 0 {
		state.Secrets, diags = types.MapValueFrom(ctx, types.StringType, job.Secrets)
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

	diags = req.State.Get(ctx,&state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := datahub.UpdateJobRequest{}
	if !plan.Name.Equal(state.Name) {
		updateReq.Name = plan.Name.ValueString()
	}

	if !plan.Type.Equal(state.Type) {
		updateReq.Type = plan.Type.ValueString()
	}

	if !plan.Image.Equal(state.Image) {
		updateReq.Image = plan.Type.ValueString()
	}	

	if !plan.Environment.Equal(state.Environment){
		var environment Environment
		diags = plan.Environment.ElementsAs(ctx, &environment, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		} 

		updateReq.Environment = environment

	}

	if !plan.Secrets.Equal(state.Secrets) {
		var secrets Secrets
		diags = plan.Secrets.ElementsAs(ctx, &secrets, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		} 

		updateReq.Secrets = secrets
	}
	
	if !plan.Command.Equal(state.Command) {
		var command Command
		diags = plan.Command.ElementsAs(ctx, &command, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.Command = command
	}
	
	_, err := r.client.UpdateJob(ctx, state.JobId.ValueString(), updateReq)
	if err != nil{
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
	// // Retrieve values from state
	// 	var state jobResourceModel
	// 	diags := req.State.Get(ctx, &state)
	// 	resp.Diagnostics.Append(diags...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}

	// 	// Delete existing job
	// 	err := r.client.DeleteJob(state.ID.ValueString())
	// 	if err != nil {
	// 		resp.Diagnostics.AddError(
	// 			"Error Deleting HashiCups Job",
	// 			"Could not delete job, unexpected error: "+err.Error(),
	// 		)
	// 		return
	// 	}
}

func (r *jobResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*datahub.DatahubClient)
}

func (r *jobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type jobResourceModel struct {
	JobId       types.String `tfsdk:"job_id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Image       types.String `tfsdk:"image"`
	Environment types.Map    `tfsdk:"environment"`
	Secrets     types.Map    `tfsdk:"secrets"`
	Command     types.List   `tfsdk:"command"`
}

// jobItemModel maps job item data.
// type jobItemModel struct {
// 	Coffee   jobItemCoffeeModel `tfsdk:"coffee"`
// 	Quantity types.Int64        `tfsdk:"quantity"`
// }

// // jobItemCoffeeModel maps coffee job item data.
// type jobItemCoffeeModel struct {
// 	ID          types.Int64   `tfsdk:"id"`
// 	Name        types.String  `tfsdk:"name"`
// 	Teaser      types.String  `tfsdk:"teaser"`
// 	Description types.String  `tfsdk:"description"`
// 	Price       types.Float64 `tfsdk:"price"`
// 	Image       types.String  `tfsdk:"image"`
// }
