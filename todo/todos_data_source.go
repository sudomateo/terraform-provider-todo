package todo

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/sudomateo/todo/todo"
)

// Compile-time assertions that our concrete todosDataSource implements the
// necessary interfaces for a data source.
var (
	_ datasource.DataSource              = &todosDataSource{}
	_ datasource.DataSourceWithConfigure = &todosDataSource{}
)

// NewTodosDataSource returns our implementation of this data source.
func NewTodosDataSource() datasource.DataSource {
	return &todosDataSource{}
}

// todosDataSource is the concrete type that implements the DataSource
// interface.
type todosDataSource struct {
	client *todo.Client
}

// todosDataSourceModel maps data source schema data to a native Go type.
type todosDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Todos []todosModel `tfsdk:"todos"`
}

// todosModel maps data source schema data to a native Go type.
type todosModel struct {
	ID          types.String `tfsdk:"id"`
	Text        types.String `tfsdk:"text"`
	Priority    types.String `tfsdk:"priority"`
	Completed   types.Bool   `tfsdk:"completed"`
	TimeCreated types.String `tfsdk:"time_created"`
	TimeUpdated types.String `tfsdk:"time_updated"`
}

// Metadata returns the data source type name.
func (d *todosDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_todos"
}

// Schema defines the configuration for the data source block.
func (d *todosDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"todos": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"text": schema.StringAttribute{
							Computed: true,
						},
						"priority": schema.StringAttribute{
							Computed: true,
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
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *todosDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state todosDataSourceModel

	todos, err := d.client.ListTodos()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read todos",
			err.Error(),
		)
		return
	}

	// Map response body to the schema and populate computed attributes.
	for _, todo := range todos {
		todostate := todosModel{
			ID:          types.StringValue(todo.ID.String()),
			Text:        types.StringValue(todo.Text),
			Priority:    types.StringValue(string(todo.Priority)),
			Completed:   types.BoolValue(todo.Completed),
			TimeCreated: types.StringValue(todo.TimeCreated.String()),
			TimeUpdated: types.StringValue(todo.TimeUpdated.String()),
		}

		state.Todos = append(state.Todos, todostate)
	}

	// Set the data source ID to a placeholder value for testing.
	state.ID = types.StringValue("todos_id_placeholder")

	// Set the state with the values from the read operation.
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *todosDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*todo.Client)
}
