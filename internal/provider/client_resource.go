package provider

import (
	"context"
	"time"

	"dev.azure.com/AllYourBI/Datahub/_git/go-datahub-sdk.git/pkg/datahub"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &clientResource{}
	_ resource.ResourceWithConfigure   = &clientResource{}
	_ resource.ResourceWithImportState = &clientResource{}
)

// NewClientResource is a helper function to simplify the provider implementation.
func NewClientResource() resource.Resource {
	return &clientResource{}
}

// clientResource is the resource implementation.
type clientResource struct {
	client *datahub.DatahubClient
}

// Metadata returns the resource type name.
func (r *clientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

// Schema defines the schema for the resource.
func (r *clientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an AYBI Datahub Client.",
		Attributes: map[string]schema.Attribute{
			"customer_code": schema.StringAttribute{
				Description: "Customer code for the client.",
				Required:    true,
			},
			"customer_name": schema.StringAttribute{
				Description: "Name of the client.",
				Required:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "ID of the client.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret": schema.StringAttribute{
				Description: "Secret of the client.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expiration_date": schema.StringAttribute{
				Description: "Expiration date of the client.",
				Computed:    false,
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *clientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var client clientResourceModel
	diags := req.Plan.Get(ctx, &client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createRequest := datahub.ClientRequest{
		CustomerCode: client.CustomerCode.ValueString(),
		CustomerName: client.CustomerName.ValueString(),
	}

	createdClient, err := r.client.Auth.CreateClient(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating client",
			"Could not create client, unexpected error: "+err.Error(),
		)
		return
	}

	client.ClientID = types.StringValue(createdClient.ClientID)
	client.ClientSecret = types.StringValue(createdClient.ClientSecret)
	client.CustomerCode = types.StringValue(createdClient.CustomerCode)
	client.CustomerName = types.StringValue(createdClient.CustomerName)
	if createdClient.ExpirationDate != nil {
		client.ExpirationDate = types.StringValue(createdClient.ExpirationDate.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		client.ExpirationDate = types.StringNull()
	}

	diags = resp.State.Set(ctx, client)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *clientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state clientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	uuidClientID, err := uuid.Parse(state.ClientID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datahub Client",
			"Could not parse client ID "+state.ClientID.ValueString()+": "+err.Error(),
		)
		return
	}

	client, err := r.client.Auth.GetClient(ctx, uuidClientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datahub Client",
			"Could not read Datahub client ID "+state.ClientID.ValueString()+": "+err.Error(),
		)
		return
	}


	state.CustomerCode = types.StringValue(client.CustomerCode)
	state.CustomerName = types.StringValue(client.CustomerName)
	if client.ExpirationDate != nil {
		state.ExpirationDate = types.StringValue(client.ExpirationDate.Format("2006-01-02T15:04:05Z07:00"))
	} else {
		state.ExpirationDate = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *clientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan clientResourceModel
	var state clientResourceModel
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

	uuidClientID, err := uuid.Parse(state.ClientID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Datahub Client",
			"Could not parse client ID "+state.ClientID.ValueString()+": "+err.Error(),
		)
		return
	}

	updateRequest := datahub.ClientRequest{
		CustomerCode:   plan.CustomerCode.ValueString(),
		CustomerName:           plan.CustomerName.ValueString(),
		
	}

	if plan.ExpirationDate.IsNull() {
		updateRequest.ExpirationDate = nil
	} else {
		expirationDate, err := time.Parse("2006-01-02T15:04:05Z07:00", plan.ExpirationDate.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Datahub Client",
				"Could not parse expiration date "+plan.ExpirationDate.ValueString()+": "+err.Error(),
			)
			return
		}
		updateRequest.ExpirationDate = &expirationDate
	}

	_, err = r.client.Auth.UpdateClient(ctx, uuidClientID, updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Datahub Client",
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
func (r *clientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state clientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// uuidClientID, err := uuid.Parse(state.ClientID.ValueString())
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Deleting Datahub Client",
	// 		"Could not parse client ID "+state.ClientID.ValueString()+": "+err.Error(),
	// 	)
	// 	return
	// }

	// err = r.client.Auth.DeleteClient(ctx, uuidClientID)
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Error Deleting Datahub Client",
	// 		"Could not delete client, unexpected error: "+err.Error(),
	// 	)
	// 	return
	// }
}

func (r *clientResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*datahub.DatahubClient)
}

func (r *clientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("client_id"), req, resp)
}

type clientResourceModel struct {
	CustomerCode   types.String `tfsdk:"customer_code"`
	CustomerName   types.String `tfsdk:"customer_name"`
	ClientID       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	ExpirationDate types.String `tfsdk:"expiration_date"`
}
