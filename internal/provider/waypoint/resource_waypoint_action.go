// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waypoint

import (
	"context"
	"encoding/base64"
	"fmt"

	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/client/waypoint_service"
	waypoint_models "github.com/hashicorp/hcp-sdk-go/clients/cloud-waypoint-service/preview/2024-11-22/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-hcp/internal/clients"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ActionResource{}
var _ resource.ResourceWithImportState = &ActionResource{}

func NewActionResource() resource.Resource {
	return &ActionResource{}
}

type ActionResource struct {
	client *clients.Client
}

// ActionResourceModel describes the resource data model.
type ActionResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ProjectID   types.String `tfsdk:"project_id"`
	OrgID       types.String `tfsdk:"organization_id"`
	Description types.String `tfsdk:"description"`

	Request *actionRequest `tfsdk:"request"`
}

type actionRequest struct {
	Custom *customRequest `tfsdk:"custom"`
	Agent  *agentRequest  `tfsdk:"agent"`
}

type customRequest struct {
	Method  types.String `tfsdk:"method"`
	Headers types.Map    `tfsdk:"headers"`
	URL     types.String `tfsdk:"url"`
	Body    types.String `tfsdk:"body"`
}

type agentRequest struct {
	OperationID types.String `tfsdk:"operation_id"`
	Body        types.String `tfsdk:"body"`
	ActionRunID types.String `tfsdk:"action_run_id"`
	Group       types.String `tfsdk:"group"`
}

func (r *ActionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_waypoint_action"
}

func (r *ActionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The Waypoint Action resource manages the lifecycle of an Action.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the Action.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the Action.",
				Required:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The ID of the HCP project where the Action is located.",
				Computed:    true,
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the HCP organization where the Action is located.",
				Computed:    true,
			},
			// An Action description must be fewer than 125 characters if set.
			"description": schema.StringAttribute{
				Description: "A description of the Action.",
				Optional:    true,
			},
			"request": schema.SingleNestedAttribute{
				Description: "The kind of HTTP request this should trigger.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"custom": schema.SingleNestedAttribute{
						Description: "Custom mode allows users to define the HTTP method, the request body, etc.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"method": schema.StringAttribute{
								Description: "The HTTP method to use for the request. Must be one of: 'GET', 'POST', 'PUT', 'DELETE', or 'PATCH'.",
								Required:    true,
								Validators: []validator.String{
									stringvalidator.OneOf("GET", "POST", "PUT", "DELETE", "PATCH"),
								},
							},
							"headers": schema.MapAttribute{
								Description: "Key value headers to send with the request.",
								Optional:    true,
								ElementType: types.StringType,
							},
							"url": schema.StringAttribute{
								Description: "The full URL this request should make when invoked.",
								Optional:    true,
							},
							"body": schema.StringAttribute{
								Description: "The body to be submitted with the request.",
								Optional:    true,
							},
						},
					},
					"agent": schema.SingleNestedAttribute{
						Description: "Agent mode allows users to define the agent to use for the request.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"operation_id": schema.StringAttribute{
								Description: "The identifying name of the operation in the agent config file.",
								Required:    true,
							},
							"body": schema.StringAttribute{
								Description: "Arguments to the operation, specified as JSON.",
								Optional:    true,
							},
							"action_run_id": schema.StringAttribute{
								Description: "An optional action run id. If specified the agent will interact with the actions subsystem.",
								Optional:    true,
							},
							"group": schema.StringAttribute{
								Description: "The name of the group that the operation is in.",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *ActionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*clients.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *ActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *ActionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID

	modelBody := &waypoint_models.HashicorpCloudWaypointV20241122WaypointServiceCreateActionConfigBody{
		ActionConfig: &waypoint_models.HashicorpCloudWaypointV20241122ActionConfig{
			Request: &waypoint_models.HashicorpCloudWaypointV20241122ActionConfigRequest{},
		},
	}

	modelBody.ActionConfig.Name = plan.Name.ValueString()

	if !plan.Description.IsUnknown() {
		modelBody.ActionConfig.Description = plan.Description.ValueString()
	}

	var diags diag.Diagnostics

	// This is a proxy for the request type, as Custom.Method is required for Custom requests
	if plan.Request.Custom != nil && !plan.Request.Custom.Method.IsUnknown() && !plan.Request.Custom.Method.IsNull() {
		modelBody.ActionConfig.Request.Custom = &waypoint_models.HashicorpCloudWaypointV20241122ActionConfigFlavorCustom{}

		method := waypoint_models.HashicorpCloudWaypointV20241122ActionConfigFlavorCustomMethod(plan.Request.Custom.Method.ValueString())
		modelBody.ActionConfig.Request.Custom.Method = &method

		if !plan.Request.Custom.Headers.IsUnknown() && !plan.Request.Custom.Headers.IsNull() {
			elements := make(map[string]types.String, len(plan.Request.Custom.Headers.Elements()))
			diags = plan.Request.Custom.Headers.ElementsAs(ctx, &elements, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			for key, value := range elements {
				modelBody.ActionConfig.Request.Custom.Headers = append(modelBody.ActionConfig.Request.Custom.Headers, &waypoint_models.HashicorpCloudWaypointV20241122ActionConfigFlavorCustomHeader{
					Key:   key,
					Value: value.ValueString(),
				})
			}
		}
		if !plan.Request.Custom.URL.IsUnknown() && !plan.Request.Custom.URL.IsNull() {
			modelBody.ActionConfig.Request.Custom.URL = plan.Request.Custom.URL.ValueString()
		}
		if !plan.Request.Custom.Body.IsUnknown() && !plan.Request.Custom.Body.IsNull() {
			modelBody.ActionConfig.Request.Custom.Body = plan.Request.Custom.Body.ValueString()
		}
	} else if plan.Request.Agent != nil && !plan.Request.Agent.OperationID.IsUnknown() && !plan.Request.Agent.OperationID.IsNull() {
		modelBody.ActionConfig.Request.Agent = &waypoint_models.HashicorpCloudWaypointV20241122ActionConfigFlavorAgent{
			Op: &waypoint_models.HashicorpCloudWaypointV20241122AgentOperation{},
		}
		if !plan.Request.Agent.OperationID.IsUnknown() && !plan.Request.Agent.OperationID.IsNull() {
			modelBody.ActionConfig.Request.Agent.Op.ID = plan.Request.Agent.OperationID.ValueString()
		}
		if !plan.Request.Agent.Group.IsUnknown() && !plan.Request.Agent.Group.IsNull() {
			modelBody.ActionConfig.Request.Agent.Op.Group = plan.Request.Agent.Group.ValueString()
		}

		if !plan.Request.Agent.Body.IsUnknown() && !plan.Request.Agent.Body.IsNull() {
			// The body is expected to be a base64 encoded string, so we decode it
			bodyBytes, err := base64.StdEncoding.DecodeString(plan.Request.Agent.Body.ValueString())
			// If there is an error, we immediately return an error to the user
			if err != nil {
				resp.Diagnostics.AddError(
					"Error decoding Agent Body",
					fmt.Sprintf("The Agent Body must be a base64 encoded string, got: %q", plan.Request.Agent.Body.ValueString()),
				)
				return
			}
			modelBody.ActionConfig.Request.Agent.Op.Body = bodyBytes
		}
		if !plan.Request.Agent.ActionRunID.IsUnknown() && !plan.Request.Agent.ActionRunID.IsNull() {
			modelBody.ActionConfig.Request.Agent.Op.ActionRunID = plan.Request.Agent.ActionRunID.ValueString()
		}
	}

	params := &waypoint_service.WaypointServiceCreateActionConfigParams{
		NamespaceLocationOrganizationID: orgID,
		NamespaceLocationProjectID:      projectID,
		Body:                            modelBody,
	}

	aCfg, err := r.client.Waypoint.WaypointServiceCreateActionConfig(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Action", err.Error())
		return
	}

	var aCfgModel *waypoint_models.HashicorpCloudWaypointV20241122ActionConfig
	if aCfg.Payload != nil {
		aCfgModel = aCfg.Payload.ActionConfig
	} else {
		resp.Diagnostics.AddError("Error creating Action", "The response payload was nil.")
		return
	}

	if aCfgModel == nil {
		resp.Diagnostics.AddError("Unknown error creating Action", "Empty Action returned")
		return
	}

	if aCfgModel.ID != "" {
		plan.ID = types.StringValue(aCfgModel.ID)
	}
	if aCfgModel.Name != "" {
		plan.Name = types.StringValue(aCfgModel.Name)
	}
	if aCfgModel.Description != "" {
		plan.Description = types.StringValue(aCfgModel.Description)
	} else {
		plan.Description = types.StringNull()
	}

	plan.ProjectID = types.StringValue(projectID)
	plan.OrgID = types.StringValue(orgID)

	plan.Request = &actionRequest{}

	if aCfgModel.Request.Custom != nil {
		diags = readCustomAction(ctx, plan, aCfgModel)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else if aCfgModel.Request.Agent != nil {
		diags = readAgentAction(ctx, plan, aCfgModel)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Created Action resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ActionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID
	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: orgID,
		ProjectID:      projectID,
	}

	client := r.client

	actionCfg, err := clients.GetAction(ctx, client, loc, data.ID.ValueString(), data.Name.ValueString())
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Action not found for organization, removing from state.")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading Action", err.Error())
		return
	}

	if actionCfg.ID != "" {
		data.ID = types.StringValue(actionCfg.ID)
	}
	if actionCfg.Name != "" {
		data.Name = types.StringValue(actionCfg.Name)
	}
	if actionCfg.Description != "" {
		data.Description = types.StringValue(actionCfg.Description)
	} else {
		data.Description = types.StringNull()
	}

	data.ProjectID = types.StringValue(projectID)
	data.OrgID = types.StringValue(orgID)

	data.Request = &actionRequest{}

	if actionCfg.Request.Custom != nil {
		diags := readCustomAction(ctx, data, actionCfg)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else if actionCfg.Request.Agent != nil {
		diags := readAgentAction(ctx, data, actionCfg)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *ActionResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// get the current state as well, so we know the current name of the
	// action for reference during the update
	var data *ActionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	projectID := r.client.Config.ProjectID
	if !plan.ProjectID.IsUnknown() {
		projectID = plan.ProjectID.ValueString()
	}

	orgID := r.client.Config.OrganizationID

	modelBody := &waypoint_models.HashicorpCloudWaypointV20241122WaypointServiceUpdateActionConfigBody{
		ActionConfig: &waypoint_models.HashicorpCloudWaypointV20241122ActionConfig{
			Request: &waypoint_models.HashicorpCloudWaypointV20241122ActionConfigRequest{},
		},
	}

	// These are the updated values
	modelBody.ActionConfig.Name = plan.Name.ValueString()
	modelBody.ActionConfig.Description = plan.Description.ValueString()

	var diags diag.Diagnostics

	// This is a proxy for the request type, as Custom.Method is required for Custom requests
	if plan.Request.Custom != nil && !plan.Request.Custom.Method.IsUnknown() && !plan.Request.Custom.Method.IsNull() {
		modelBody.ActionConfig.Request.Custom = &waypoint_models.HashicorpCloudWaypointV20241122ActionConfigFlavorCustom{}

		method := waypoint_models.HashicorpCloudWaypointV20241122ActionConfigFlavorCustomMethod(plan.Request.Custom.Method.ValueString())
		modelBody.ActionConfig.Request.Custom.Method = &method

		if !plan.Request.Custom.Headers.IsUnknown() && !plan.Request.Custom.Headers.IsNull() {
			elements := make(map[string]types.String, len(plan.Request.Custom.Headers.Elements()))
			diags = plan.Request.Custom.Headers.ElementsAs(ctx, &elements, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			for key, value := range elements {
				modelBody.ActionConfig.Request.Custom.Headers = append(modelBody.ActionConfig.Request.Custom.Headers, &waypoint_models.HashicorpCloudWaypointV20241122ActionConfigFlavorCustomHeader{
					Key:   key,
					Value: value.ValueString(),
				})
			}
		}
		if !plan.Request.Custom.URL.IsUnknown() && !plan.Request.Custom.URL.IsNull() {
			modelBody.ActionConfig.Request.Custom.URL = plan.Request.Custom.URL.ValueString()
		}
		if !plan.Request.Custom.Body.IsUnknown() && !plan.Request.Custom.Body.IsNull() {
			modelBody.ActionConfig.Request.Custom.Body = plan.Request.Custom.Body.ValueString()

		}
		// This is a proxy for the request type, as Agent.OperationID is required for Agent requests
	} else if plan.Request.Agent != nil && !plan.Request.Agent.OperationID.IsUnknown() && !plan.Request.Agent.OperationID.IsNull() {
		modelBody.ActionConfig.Request.Agent = &waypoint_models.HashicorpCloudWaypointV20241122ActionConfigFlavorAgent{
			Op: &waypoint_models.HashicorpCloudWaypointV20241122AgentOperation{},
		}

		if !plan.Request.Agent.OperationID.IsUnknown() && !plan.Request.Agent.OperationID.IsNull() {
			modelBody.ActionConfig.Request.Agent.Op.ID = plan.Request.Agent.OperationID.ValueString()
		}
		if !plan.Request.Agent.Group.IsUnknown() && !plan.Request.Agent.Group.IsNull() {
			modelBody.ActionConfig.Request.Agent.Op.Group = plan.Request.Agent.Group.ValueString()
		}

		if !plan.Request.Agent.Body.IsUnknown() && !plan.Request.Agent.Body.IsNull() {
			// The body is expected to be a base64 encoded string, so we decode it
			bodyBytes, err := base64.StdEncoding.DecodeString(plan.Request.Agent.Body.ValueString())
			// If there is an error, we immediately return an error to the user
			if err != nil {
				resp.Diagnostics.AddError(
					"Error decoding Agent Body",
					fmt.Sprintf("The Agent Body must be a base64 encoded string, got: %q", plan.Request.Agent.Body.ValueString()),
				)
				return
			} else {
				modelBody.ActionConfig.Request.Agent.Op.Body = bodyBytes
			}
		}
		if !plan.Request.Agent.ActionRunID.IsUnknown() && !plan.Request.Agent.ActionRunID.IsNull() {
			modelBody.ActionConfig.Request.Agent.Op.ActionRunID = plan.Request.Agent.ActionRunID.ValueString()
		}
	}

	params := &waypoint_service.WaypointServiceUpdateActionConfigParams{
		NamespaceLocationOrganizationID: orgID,
		NamespaceLocationProjectID:      projectID,
		Body:                            modelBody,
	}

	actionCfg, err := r.client.Waypoint.WaypointServiceUpdateActionConfig(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error updating Action", err.Error())
		return
	}

	var aCfgModel *waypoint_models.HashicorpCloudWaypointV20241122ActionConfig
	if actionCfg.Payload != nil {
		aCfgModel = actionCfg.Payload.ActionConfig
	}
	if aCfgModel == nil {
		resp.Diagnostics.AddError("Unknown error updating Action", "Empty Action returned")
		return
	}

	if aCfgModel.ID != "" {
		plan.ID = types.StringValue(aCfgModel.ID)
	}
	if aCfgModel.Name != "" {
		plan.Name = types.StringValue(aCfgModel.Name)
	}
	if aCfgModel.Description != "" {
		plan.Description = types.StringValue(aCfgModel.Description)
	} else {
		plan.Description = types.StringNull()
	}

	plan.ProjectID = types.StringValue(projectID)
	plan.OrgID = types.StringValue(orgID)

	plan.Request = &actionRequest{}

	if aCfgModel.Request.Custom != nil {
		diags = readCustomAction(ctx, plan, aCfgModel)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else if aCfgModel.Request.Agent != nil {
		diags = readAgentAction(ctx, plan, aCfgModel)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "Updated Action resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ActionResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	projectID := r.client.Config.ProjectID
	if !data.ProjectID.IsUnknown() {
		projectID = data.ProjectID.ValueString()
	}

	loc := &sharedmodels.HashicorpCloudLocationLocation{
		OrganizationID: r.client.Config.OrganizationID,
		ProjectID:      projectID,
	}

	params := &waypoint_service.WaypointServiceDeleteActionConfigParams{
		NamespaceLocationOrganizationID: loc.OrganizationID,
		NamespaceLocationProjectID:      loc.ProjectID,
		ActionID:                        data.ID.ValueStringPointer(),
		ActionName:                      data.Name.ValueStringPointer(),
	}

	_, err := r.client.Waypoint.WaypointServiceDeleteActionConfig(params, nil)
	if err != nil {
		if clients.IsResponseCodeNotFound(err) {
			tflog.Info(ctx, "Action not found for organization during delete call, ignoring")
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting Action",
			err.Error(),
		)
		return
	}
}

func (r *ActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func readCustomAction(
	ctx context.Context,
	data *ActionResourceModel,
	actionCfg *waypoint_models.HashicorpCloudWaypointV20241122ActionConfig,
) diag.Diagnostics {
	data.Request.Custom = &customRequest{}
	headerMap := make(map[string]string)
	var diags diag.Diagnostics
	if actionCfg.Request.Custom.Method != nil {
		data.Request.Custom.Method = types.StringValue(string(*actionCfg.Request.Custom.Method))
	} else {
		data.Request.Custom.Method = types.StringNull()
	}
	if actionCfg.Request.Custom.Headers != nil {
		for _, header := range actionCfg.Request.Custom.Headers {
			headerMap[header.Key] = header.Value
		}
		if len(headerMap) > 0 {
			data.Request.Custom.Headers, diags = types.MapValueFrom(ctx, types.StringType, headerMap)
			if diags.HasError() {
				return diags
			}
		} else {
			data.Request.Custom.Headers = types.MapNull(types.StringType)
		}
	}
	if actionCfg.Request.Custom.URL != "" {
		data.Request.Custom.URL = types.StringValue(actionCfg.Request.Custom.URL)
	} else {
		data.Request.Custom.URL = types.StringNull()
	}
	if actionCfg.Request.Custom.Body != "" {
		data.Request.Custom.Body = types.StringValue(actionCfg.Request.Custom.Body)
	} else {
		data.Request.Custom.Body = types.StringNull()
	}

	// Ensure Agent is nil if Custom is set
	data.Request.Agent = nil

	return diags
}

func readAgentAction(
	ctx context.Context,
	data *ActionResourceModel,
	actionCfg *waypoint_models.HashicorpCloudWaypointV20241122ActionConfig,
) diag.Diagnostics {
	data.Request.Agent = &agentRequest{}
	var diags diag.Diagnostics
	if actionCfg.Request.Agent.Op.ID != "" {
		data.Request.Agent.OperationID = types.StringValue(actionCfg.Request.Agent.Op.ID)
	} else {
		data.Request.Agent.OperationID = types.StringNull()
	}
	if actionCfg.Request.Agent.Op.Group != "" {
		data.Request.Agent.Group = types.StringValue(actionCfg.Request.Agent.Op.Group)
	} else {
		data.Request.Agent.Group = types.StringNull()
	}
	if actionCfg.Request.Agent.Op.Body.String() != "" {
		data.Request.Agent.Body = types.StringValue(base64.StdEncoding.EncodeToString(actionCfg.Request.Agent.Op.Body))
	} else {
		data.Request.Agent.Body = types.StringNull()
	}
	if actionCfg.Request.Agent.Op.ActionRunID != "" {
		data.Request.Agent.ActionRunID = types.StringValue(actionCfg.Request.Agent.Op.ActionRunID)
	} else {
		data.Request.Agent.ActionRunID = types.StringNull()
	}

	// Ensure Custom is nil if Agent is set
	data.Request.Custom = nil

	return diags
}
