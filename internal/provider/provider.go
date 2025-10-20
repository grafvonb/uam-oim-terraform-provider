package provider

import (
	"context"
	"os"

	"github.com/hashicorp-demoapp/hashicups-client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &uamoimProvider{}
)

type uamoimProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type uamoimProviderConfig struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &uamoimProvider{
			version: version,
		}
	}
}

func (p *uamoimProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "uamoim"
	resp.Version = p.version
}

func (p *uamoimProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *uamoimProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var cfg uamoimProviderConfig
	diags := req.Config.Get(ctx, &cfg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if cfg.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown uamoim API Host",
			"The provider cannot create the uamoim API client as there is an unknown configuration value for the uamoim API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the uamoim_HOST environment variable.",
		)
	}
	if cfg.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown uamoim API Username",
			"The provider cannot create the uamoim API client as there is an unknown configuration value for the uamoim API username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the uamoim_USERNAME environment variable.",
		)
	}
	if cfg.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown uamoim API Password",
			"The provider cannot create the uamoim API client as there is an unknown configuration value for the uamoim API password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the uamoim_PASSWORD environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("UAMOIM_HOST")
	username := os.Getenv("UAMOIM_USERNAME")
	password := os.Getenv("UAMOIM_PASSWORD")
	if !cfg.Host.IsNull() {
		host = cfg.Host.ValueString()
	}
	if !cfg.Username.IsNull() {
		username = cfg.Username.ValueString()
	}
	if !cfg.Password.IsNull() {
		password = cfg.Password.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing uamoim API Host",
			"The provider cannot create the uamoim API client as there is a missing or empty value for the uamoim API host. "+
				"Set the host value in the configuration or use the uamoim_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing uamoim API Username",
			"The provider cannot create the uamoim API client as there is a missing or empty value for the uamoim API username. "+
				"Set the username value in the configuration or use the uamoim_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing uamoim API Password",
			"The provider cannot create the uamoim API client as there is a missing or empty value for the uamoim API password. "+
				"Set the password value in the configuration or use the uamoim_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new uamoim client using the configuration values
	client, err := hashicups.NewClient(&host, &username, &password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create uamoim API Client",
			"An unexpected error occurred when creating the uamoim API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"uamoim Client Error: "+err.Error(),
		)
		return
	}

	// Make the uamoim client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *uamoimProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewShopsDataSource, NewSODsDataSource, NewCoffeesDataSource,
	}
}

func (p *uamoimProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
