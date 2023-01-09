package terraform

import (
	"context"
	// "fmt"

	// "strconv"
	"terraform-provider-aybi-datahub/internal/datahub"
	// "time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	// "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	// "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
			// "job_id": schema.StringAttribute{
			// 	Description: "Numeric identifier of the job.",
			// 	Computed:    true,
			// 	PlanModifiers: []planmodifier.String{
			// 		stringplanmodifier.UseStateForUnknown(),
			// 	},
			// },
			"image": schema.StringAttribute{
				Description: "Docker image for the job",
				Required:    true,
			},
			// "environment": schema.MapAttribute{
			// 	ElementType: types.StringType,
			// 	Optional: true,
			// },
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

	// tflog.Debug(ctx, fmt.Sprintf("Got environment variables %v", job.Environment.Elements()))

	// // Generate API request body from plan
	// var items []hashicups.JobItem
	// for _, item := range plan.Items {
	//     items = append(items, hashicups.JobItem{
	//         Coffee: hashicups.Coffee{
	//             ID: int(item.Coffee.ID.ValueInt64()),
	//         },
	//         Quantity: int(item.Quantity.ValueInt64()),
	//     })
	// }

	// // Create new job
	// job, err := r.client.CreateJob(items)
	// if err != nil {
	//     resp.Diagnostics.AddError(
	//         "Error creating job",
	//         "Could not create job, unexpected error: "+err.Error(),
	//     )
	//     return
	// }

	// // Map response body to schema and populate Computed attribute values
	// plan.ID = types.StringValue(strconv.Itoa(job.ID))
	// for jobItemIndex, jobItem := range job.Items {
	//     plan.Items[jobItemIndex] = jobItemModel{
	//         Coffee: jobItemCoffeeModel{
	//             ID:          types.Int64Value(int64(jobItem.Coffee.ID)),
	//             Name:        types.StringValue(jobItem.Coffee.Name),
	//             Teaser:      types.StringValue(jobItem.Coffee.Teaser),
	//             Description: types.StringValue(jobItem.Coffee.Description),
	//             Price:       types.Float64Value(jobItem.Coffee.Price),
	//             Image:       types.StringValue(jobItem.Coffee.Image),
	//         },
	//         Quantity: types.Int64Value(int64(jobItem.Quantity)),
	//     }
	// }
	// plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// // Set state to fully populated data
	// diags = resp.State.Set(ctx, plan)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	//     return
	// }
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
	// 	job, err := r.client.GetJob(state.ID.ValueString())
	// 	if err != nil {
	// 		resp.Diagnostics.AddError(
	// 			"Error Reading HashiCups Job",
	// 			"Could not read HashiCups job ID "+state.ID.ValueString()+": "+err.Error(),
	// 		)
	// 		return
	// 	}

	// 	// Overwrite items with refreshed state
	// 	state.Items = []jobItemModel{}
	// 	for _, item := range job.Items {
	// 		state.Items = append(state.Items, jobItemModel{
	// 			Coffee: jobItemCoffeeModel{
	// 				ID:          types.Int64Value(int64(item.Coffee.ID)),
	// 				Name:        types.StringValue(item.Coffee.Name),
	// 				Teaser:      types.StringValue(item.Coffee.Teaser),
	// 				Description: types.StringValue(item.Coffee.Description),
	// 				Price:       types.Float64Value(item.Coffee.Price),
	// 				Image:       types.StringValue(item.Coffee.Image),
	// 			},
	// 			Quantity: types.Int64Value(int64(item.Quantity)),
	// 		})
	// 	}

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
	// var plan jobResourceModel
	// diags := req.Plan.Get(ctx, &plan)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	//     return
	// }

	// // Generate API request body from plan
	// var hashicupsItems []hashicups.JobItem
	// for _, item := range plan.Items {
	//     hashicupsItems = append(hashicupsItems, hashicups.JobItem{
	//         Coffee: hashicups.Coffee{
	//             ID: int(item.Coffee.ID.ValueInt64()),
	//         },
	//         Quantity: int(item.Quantity.ValueInt64()),
	//     })
	// }

	// // Update existing job
	// _, err := r.client.UpdateJob(plan.ID.ValueString(), hashicupsItems)
	// if err != nil {
	//     resp.Diagnostics.AddError(
	//         "Error Updating HashiCups Job",
	//         "Could not update job, unexpected error: "+err.Error(),
	//     )
	//     return
	// }

	// // Fetch updated items from GetJob as UpdateJob items are not
	// // populated.
	// job, err := r.client.GetJob(plan.ID.ValueString())
	// if err != nil {
	//     resp.Diagnostics.AddError(
	//         "Error Reading HashiCups Job",
	//         "Could not read HashiCups job ID "+plan.ID.ValueString()+": "+err.Error(),
	//     )
	//     return
	// }

	// // Update resource state with updated items and timestamp
	// plan.Items = []jobItemModel{}
	// for _, item := range job.Items {
	//     plan.Items = append(plan.Items, jobItemModel{
	//         Coffee: jobItemCoffeeModel{
	//             ID:          types.Int64Value(int64(item.Coffee.ID)),
	//             Name:        types.StringValue(item.Coffee.Name),
	//             Teaser:      types.StringValue(item.Coffee.Teaser),
	//             Description: types.StringValue(item.Coffee.Description),
	//             Price:       types.Float64Value(item.Coffee.Price),
	//             Image:       types.StringValue(item.Coffee.Image),
	//         },
	//         Quantity: types.Int64Value(int64(item.Quantity)),
	//     })
	// }
	// plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// diags = resp.State.Set(ctx, plan)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	//     return
	// }
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
	// JobId       types.String `tfsdk:"job_id"`
	Image       types.String `tfsdk:"image"`
	// Environment types.Map    `tfsdk:"environment"`
}

// jobItemModel maps job item data.
type jobItemModel struct {
	Coffee   jobItemCoffeeModel `tfsdk:"coffee"`
	Quantity types.Int64        `tfsdk:"quantity"`
}

// jobItemCoffeeModel maps coffee job item data.
type jobItemCoffeeModel struct {
	ID          types.Int64   `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Teaser      types.String  `tfsdk:"teaser"`
	Description types.String  `tfsdk:"description"`
	Price       types.Float64 `tfsdk:"price"`
	Image       types.String  `tfsdk:"image"`
}
