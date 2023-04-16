package todo

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sudomateo/todo/todo"
)

// Compile-time assertions that our concrete todoProvider implements the
// Provider interface.
var (
	_ provider.Provider = &todoProvider{}
)

// New returns our implementation of this provider.
func New() provider.Provider {
	return &todoProvider{}
}

// todoProvider is the concrete type that implements the Provider interface.
type todoProvider struct{}

// todoProviderModel maps provider schema data to a native Go type.
type todoProviderModel struct {
	Host types.String `tfsdk:"host"`
}

// Metadata returns the provider type name.
func (p *todoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "todo"
}

// Schema defines the configuration for the provider block.
func (p *todoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Configure creates an API client for the todo API that will be used by
// resources and data sources.
func (p *todoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring todo client")

	// Retrieve provider data from configuration.
	var config todoProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure the host attribute is a known value.
	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown todo API host",
			"The provider cannot create the todo API client as there is an unknown configuration value for the todo API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the TODO_HOST environment variable.",
		)
	}

	// We had at least one error configuring the provider, return early.
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the host value from the environment, but override it if passed in
	// the configuration.
	host := os.Getenv("TODO_HOST")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	// We don't have a host, add an error.
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing todo API host",
			"The provider cannot create the todo API client as there is a missing or empty value for the todo API host. "+
				"Set the host value in the configuration or use the TODO_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// We had at least one error configuring the provider, return early.
	if resp.Diagnostics.HasError() {
		return
	}

	// Set fields for loggin.
	ctx = tflog.SetField(ctx, "todo_host", host)

	tflog.Debug(ctx, "Creating todo client")

	// Create a new todo client using the values from the configuration.
	client, err := todo.NewClient(host)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create todo API client",
			"An unexpected error occurred when creating the todo API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"todo client error: "+err.Error(),
		)
		return
	}

	// Make the todo client available to resources and data sources Configure
	// methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured todo client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented by this provider.
func (p *todoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTodosDataSource,
	}
}

// Resources defines the resources implemented by this provider.
func (p *todoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTodoResource,
	}
}
