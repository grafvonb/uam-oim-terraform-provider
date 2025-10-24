package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp-demoapp/hashicups-client-go"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &coffeesDataSource{}
	_ datasource.DataSourceWithConfigure = &coffeesDataSource{}
)

type coffeesDataSourceModel struct {
	Coffees []coffeesModel          `tfsdk:"coffees"`
	ByName  map[string]types.String `tfsdk:"by_name"`
	ByID    map[string]types.String `tfsdk:"by_id"`
}

type coffeesModel struct {
	ID          types.Int64               `tfsdk:"id"`
	Name        types.String              `tfsdk:"name"`
	Teaser      types.String              `tfsdk:"teaser"`
	Description types.String              `tfsdk:"description"`
	Price       types.Float64             `tfsdk:"price"`
	Image       types.String              `tfsdk:"image"`
	Ingredients []coffeesIngredientsModel `tfsdk:"ingredients"`
}
type coffeesIngredientsModel struct {
	ID types.Int64 `tfsdk:"id"`
}

func NewCoffeesDataSource() datasource.DataSource {
	return &coffeesDataSource{}
}

type coffeesDataSource struct {
	client *hashicups.Client
}

func (d *coffeesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*hashicups.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *coffeesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_coffees"
}

func (d *coffeesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"coffees": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"teaser": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"price": schema.Float64Attribute{
							Computed: true,
						},
						"image": schema.StringAttribute{
							Computed: true,
						},
						"ingredients": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.Int64Attribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"by_name": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A map of coffee names by their ID.",
			},
			"by_id": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A map of coffee IDs by their name.",
			},
		},
	}
}

func (d *coffeesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state coffeesDataSourceModel
	coffees, err := d.client.GetCoffees()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read HashiCups Coffees",
			err.Error(),
		)
		return
	}
	for _, coffee := range coffees {
		coffeeState := coffeesModel{
			ID:          types.Int64Value(int64(coffee.ID)),
			Name:        types.StringValue(coffee.Name),
			Teaser:      types.StringValue(coffee.Teaser),
			Description: types.StringValue(coffee.Description),
			Price:       types.Float64Value(coffee.Price),
			Image:       types.StringValue(coffee.Image),
		}
		for _, ingredient := range coffee.Ingredient {
			coffeeState.Ingredients = append(coffeeState.Ingredients, coffeesIngredientsModel{
				ID: types.Int64Value(int64(ingredient.ID)),
			})
		}
		state.Coffees = append(state.Coffees, coffeeState)
	}

	byName := make(map[string]types.String, len(state.Coffees))
	byID := make(map[string]types.String, len(state.Coffees))
	for _, coffee := range coffees {
		byID[fmt.Sprintf("%d", coffee.ID)] = types.StringValue(coffee.Name)
		byName[coffee.Name] = types.StringValue(fmt.Sprintf("%d", coffee.ID))
	}
	state.ByName = byName
	state.ByID = byID

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
