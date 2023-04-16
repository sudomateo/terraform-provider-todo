package todo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sudomateo/todo/todo"
)

// Compile-time assertions that our concrete todoResource implements the
// necessary interfaces for a resource.
var (
	_ resource.Resource                = &todoResource{}
	_ resource.ResourceWithConfigure   = &todoResource{}
	_ resource.ResourceWithImportState = &todoResource{}
)

// NewTodoResource returns our implementation of this resource.
func NewTodoResource() resource.Resource {
	return &todoResource{}
}

// todoResource is the concrete type that implements the Resource interface.
type todoResource struct {
	client *todo.Client
}

// todoResourceModel maps resource schema data to a native Go type.
type todoResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Text        types.String `tfsdk:"text"`
	Priority    types.String `tfsdk:"priority"`
	Completed   types.Bool   `tfsdk:"completed"`
	TimeCreated types.String `tfsdk:"time_created"`
	TimeUpdated types.String `tfsdk:"time_updated"`
}

// Metadata returns the resource type name.
func (r *todoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_todo"
}

// Schema defines the configuration for the resource block.
func (r *todoResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"text": schema.StringAttribute{
				Required: true,
			},
			"priority": schema.StringAttribute{
				Optional: true,
			},
			"completed": schema.BoolAttribute{
				Computed: true,
			},
			"time_created": schema.StringAttribute{
				Computed: true,
			},
			"time_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *todoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan.
	var plan todoResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default the priority if not provided.
	priority := plan.Priority.ValueString()
	if priority == "" {
		priority = string(todo.PriorityLow)
	}

	// Generate an API request body from retrieved plan values.
	params := todo.TodoCreateParams{
		Text:     plan.Text.ValueString(),
		Priority: todo.Priority(priority),
	}

	// Create new todo.
	td, err := r.client.CreateTodo(params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating todo",
			"Could not create todo, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to the schema and populate computed attributes.
	plan.ID = types.StringValue(td.ID.String())
	plan.Text = types.StringValue(td.Text)
	plan.Priority = types.StringValue(string(td.Priority))
	plan.Completed = types.BoolValue(td.Completed)
	plan.TimeCreated = types.StringValue(td.TimeCreated.String())
	plan.TimeUpdated = types.StringValue(td.TimeUpdated.String())

	// Set the state with the values from the create operation.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *todoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state.
	var state todoResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed todo from the API.
	td, err := r.client.GetTodo(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading todo",
			"Could not read todo ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Map response body to the schema and populate computed attributes.
	state.Text = types.StringValue(td.Text)
	state.Priority = types.StringValue(string(td.Priority))
	state.Completed = types.BoolValue(td.Completed)
	state.TimeCreated = types.StringValue(td.TimeCreated.String())
	state.TimeUpdated = types.StringValue(td.TimeUpdated.String())

	// Set the state with the values from the read operation.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *todoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan.
	var plan todoResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan.
	text := plan.Text.ValueString()
	priority := todo.Priority(plan.Priority.ValueString())
	completed := plan.Completed.ValueBool()
	params := todo.TodoUpdateParams{
		Text:      &text,
		Priority:  &priority,
		Completed: &completed,
	}

	// Update existing todo.
	td, err := r.client.UpdateTodo(plan.ID.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating todo",
			"Could not update todo, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to the schema and populate computed attributes.
	plan.Text = types.StringValue(td.Text)
	plan.Priority = types.StringValue(string(td.Priority))
	plan.Completed = types.BoolValue(td.Completed)
	plan.TimeCreated = types.StringValue(td.TimeCreated.String())
	plan.TimeUpdated = types.StringValue(td.TimeUpdated.String())

	// Set the state with the values from the update operation.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *todoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state.
	var state todoResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing todo.
	err := r.client.DeleteTodo(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting todo",
			"Could not delete todo, unexpected error: "+err.Error(),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *todoResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*todo.Client)
}

// ImportState uses a resources Read method to implement import.
func (r *todoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
