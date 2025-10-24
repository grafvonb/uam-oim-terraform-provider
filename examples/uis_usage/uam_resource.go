//go:build uis_usage

package uis_usage

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

/*
import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ModuleResource struct {
	client *api.Client
}

type ModuleModel struct {
	ID              types.String `tfsdk:"id"`
	ApplicationName types.String `tfsdk:"application_name"`
	ModuleName      types.String `tfsdk:"module_name"`
	Description     types.String `tfsdk:"description"`
}

func (r *ModuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_name": schema.StringAttribute{Required: true},
			"module_name":      schema.StringAttribute{Required: true},
			"description":      schema.StringAttribute{Optional: true},
		},
	}
}

func (r *ModuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ModuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.Modules.Create(ctx, api.ModuleCreate{
		ApplicationName: plan.ApplicationName.ValueString(),
		Name:            plan.ModuleName.ValueString(),
		Description:     plan.Description.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("create failed", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ModuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ModuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	m, err := r.client.Modules.Get(ctx, state.ID.ValueString())
	if api.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("read failed", err.Error())
		return
	}

	state.ModuleName = types.StringValue(m.Name)
	state.Description = types.StringValue(m.Description)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ModuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ModuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Modules.Update(ctx, plan.ID.ValueString(), api.ModuleUpdate{
		Name:        plan.ModuleName.ValueString(),
		Description: plan.Description.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("update failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ModuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ModuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.Modules.Delete(ctx, state.ID.ValueString()); err != nil && !api.IsNotFound(err) {
		resp.Diagnostics.AddError("delete failed", err.Error())
	}
}
