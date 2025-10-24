package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp-demoapp/hashicups-client-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64default"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource               = &orderResource{}
	_ resource.ResourceWithConfigure  = &orderResource{}
	_ resource.ResourceWithModifyPlan = &orderResource{}
)

// NewOrderResource is a helper function to simplify the provider implementation.
func NewOrderResource() resource.Resource {
	return &orderResource{}
}

// orderResource is the resource implementation.
type orderResource struct {
	client *hashicups.Client
}

func (r *orderResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan orderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() || len(plan.Items) == 0 {
		return
	}

	coffees, err := r.client.GetCoffees()
	if err != nil {
		resp.Diagnostics.AddError("Error Reading HashiCups Coffees", err.Error())
		return
	}

	// Build lookup: coffee ID -> price
	priceByID := make(map[int64]float64, len(coffees))
	for _, c := range coffees {
		priceByID[int64(c.ID)] = c.Price
	}

	for i := range plan.Items {
		// only set when price is missing/unknown
		// if !(plan.Items[i].Coffee.Price.IsNull() || plan.Items[i].Coffee.Price.IsUnknown()) {
		//continue
		//}
		// need a known ID to look up
		if plan.Items[i].Coffee.ID.IsNull() || plan.Items[i].Coffee.ID.IsUnknown() {
			continue // or add an error if ID must be provided
		}

		id := plan.Items[i].Coffee.ID.ValueInt64()
		price, ok := priceByID[id]
		if !ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("items").AtListIndex(i).AtName("coffee").AtName("id"),
				"Unknown coffee ID",
				fmt.Sprintf("ID %d not found in available coffees", id),
			)
			continue
		}

		p := path.Root("items").AtListIndex(i).AtName("coffee").AtName("price")
		if err := resp.Plan.SetAttribute(ctx, p, types.Float64Value(price)); err != nil {
			resp.Diagnostics.Append(err...)
		}
	}

	// Do not call resp.Plan.Set, Rewriting the whole plan:
	// - destroys unknown values coming from other resources
	// - clobbers results of other plan modifiers
	// - drops RequiresReplace intent and other planner metadata
	// - can introduce wrong diffs
	// diags = resp.Plan.Set(ctx, &plan)
	// resp.Diagnostics.Append(diags...)
}

func (r *orderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*hashicups.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *hashicups.Client")
		return
	}
	r.client = client
}

// Metadata returns the resource type name.
func (r *orderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_order"
}

// Schema defines the schema for the resource.
func (r *orderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
			"items": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"quantity": schema.Int64Attribute{
							Required: true,
						},
						"coffee": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"id": schema.Int64Attribute{
									Required: true,
								},
								"name": schema.StringAttribute{
									Optional: true,
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
									Optional: true,
									Default:  float64default.StaticFloat64(0.0),
								},
								"image": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// orderResourceModel maps the resource schema data.
type orderResourceModel struct {
	ID          types.String     `tfsdk:"id"`
	Items       []orderItemModel `tfsdk:"items"`
	LastUpdated types.String     `tfsdk:"last_updated"`
}

// orderItemModel maps order item data.
type orderItemModel struct {
	Coffee   orderItemCoffeeModel `tfsdk:"coffee"`
	Quantity types.Int64          `tfsdk:"quantity"`
}

// orderItemCoffeeModel maps coffee order item data.
type orderItemCoffeeModel struct {
	ID          types.Int64   `tfsdk:"id"`
	Name        types.String  `tfsdk:"name"`
	Teaser      types.String  `tfsdk:"teaser"`
	Description types.String  `tfsdk:"description"`
	Price       types.Float64 `tfsdk:"price"`
	Image       types.String  `tfsdk:"image"`
}

// Create a new resource.
// The provider uses the Create method to create a new resource based on the schema data.
// The create method follows these steps:
// 1. Checks whether the API Client is configured. If not, the resource responds with an error.
// 2. Retrieves values from the plan. The function will attempt to retrieve values from the plan and convert it to an orderResourceModel.
// 3. Generates an API request body from the plan values. The function loops through each plan item and maps it to a hashicups.OrderItem. This is what the API client needs to create a new order.
// 4. Creates a new order. The function invokes the API client's CreateOrder method.
// 5. Maps response body to resource schema attributes. After the function creates an order, it maps the hashicups.Order response to []OrderItem so the provider can update the Terraform state.
// 6. Sets Terraform's state with the new order's details.
func (r *orderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan orderResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Generate API request body from plan
	var items []hashicups.OrderItem
	for _, item := range plan.Items {
		items = append(items, hashicups.OrderItem{
			Coffee: hashicups.Coffee{
				ID: int(item.Coffee.ID.ValueInt64()),
			},
			Quantity: int(item.Quantity.ValueInt64()),
		})
	}
	// Create new order
	order, err := r.client.CreateOrder(items)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}
	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(strconv.Itoa(order.ID))
	for orderItemIndex, orderItem := range order.Items {
		plan.Items[orderItemIndex] = orderItemModel{
			Coffee: orderItemCoffeeModel{
				ID:          types.Int64Value(int64(orderItem.Coffee.ID)),
				Name:        types.StringValue(orderItem.Coffee.Name),
				Teaser:      types.StringValue(orderItem.Coffee.Teaser),
				Description: types.StringValue(orderItem.Coffee.Description),
				Price:       types.Float64Value(orderItem.Coffee.Price),
				Image:       types.StringValue(orderItem.Coffee.Image),
			},
			Quantity: types.Int64Value(int64(orderItem.Quantity)),
		}
	}
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
// The provider uses the Read function to retrieve the resource's information and update the Terraform state to reflect the resource's current state. The provider invokes this function before every plan to generate an accurate diff between the resource's current state and the configuration.
// The read function follows these steps:
// 1. Gets the current state. If it is unable to, the provider responds with an error.
// 2. Retrieves the order ID from Terraform's state.
// 3. Retrieves the order details from the client. The function invokes the API client's GetOrder method with the order ID.
// 4. Maps the response body to resource schema attributes. After the function retrieves the order, it maps the hashicups.Order response to []OrderItem so the provider can update the Terraform state.
// 5. Set Terraform's state with the order's details.
func (r *orderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state orderResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from HashiCups
	order, err := r.client.GetOrder(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading HashiCups Order",
			"Could not read HashiCups order ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.Items = []orderItemModel{}
	for _, item := range order.Items {
		state.Items = append(state.Items, orderItemModel{
			Coffee: orderItemCoffeeModel{
				ID:          types.Int64Value(int64(item.Coffee.ID)),
				Name:        types.StringValue(item.Coffee.Name),
				Teaser:      types.StringValue(item.Coffee.Teaser),
				Description: types.StringValue(item.Coffee.Description),
				Price:       types.Float64Value(item.Coffee.Price),
				Image:       types.StringValue(item.Coffee.Image),
			},
			Quantity: types.Int64Value(int64(item.Quantity)),
		})
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *orderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *orderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
