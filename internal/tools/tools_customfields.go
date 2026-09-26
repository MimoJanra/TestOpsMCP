package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
)

func customFieldDtoToMap(cf allure.CustomFieldDto) map[string]any {
	m := map[string]any{
		"id":            cf.ID,
		"name":          cf.Name,
		"archived":      cf.Archived,
		"locked":        cf.Locked,
		"required":      cf.Required,
		"single_select": cf.SingleSelect,
	}
	if cf.DefaultCustomFieldValueID != nil {
		m["default_custom_field_value_id"] = *cf.DefaultCustomFieldValueID
	}
	return m
}

func (r *Registry) registerCustomFieldTools() {
	r.register(&Tool{
		Name: "create_custom_field",
		Description: "Create a new custom field definition (org-wide). The field is not usable in any project " +
			"until it's attached with add_custom_fields_to_project, and has no selectable values until " +
			"create_custom_field_value adds some.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Field name (e.g. \"Severity\")",
				},
				"required": map[string]any{
					"type":        "boolean",
					"description": "Whether a value for this field must be set on every test case it's attached to",
				},
				"single_select": map[string]any{
					"type":        "boolean",
					"description": "Whether a test case may have at most one value for this field (optional, default false)",
				},
			},
			"required": []string{"name", "required"},
		},
		Handler: Typed(r.createCustomField),
	})

	r.register(&Tool{
		Name:        "get_custom_field",
		Description: "Get a custom field definition by ID",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"custom_field_id": map[string]any{
					"type":        "integer",
					"description": "Custom field ID",
				},
			},
			"required": []string{"custom_field_id"},
		},
		Handler: Typed(r.getCustomField),
	})

	r.register(&Tool{
		Name: "update_custom_field",
		Description: "Update a custom field definition's name, required flag, single_select flag, or locked " +
			"state. All fields are optional — only the ones you pass are changed. Note: locked does NOT block renaming the field " +
			"or adding/renaming values (confirmed live). Switching to single_select makes each later add replace a test case's value.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"custom_field_id": map[string]any{
					"type":        "integer",
					"description": "Custom field ID",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "New field name (optional)",
				},
				"required": map[string]any{
					"type":        "boolean",
					"description": "Whether a value must be set on every test case it's attached to (optional)",
				},
				"single_select": map[string]any{
					"type":        "boolean",
					"description": "Whether a test case may have at most one value for this field (optional)",
				},
				"locked": map[string]any{
					"type":        "boolean",
					"description": "Whether the field's definition is locked against further changes (optional)",
				},
			},
			"required": []string{"custom_field_id"},
		},
		Handler: Typed(r.updateCustomField),
	})

	r.register(&Tool{
		Name: "delete_custom_field",
		Description: "Permanently delete a custom field definition and all its values. The API refuses while the field is attached to any project " +
			"(custom-field.in-use.project): clear its values from test cases, remove_custom_field_from_project, then delete. Prefer set_custom_field_archived for a reversible removal.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"custom_field_id": map[string]any{
					"type":        "integer",
					"description": "Custom field ID",
				},
			},
			"required": []string{"custom_field_id"},
		},
		Handler: Typed(r.deleteCustomField),
	})

	r.register(&Tool{
		Name:        "set_custom_field_archived",
		Description: "Archive or unarchive a custom field (a reversible \"soft delete\" — an archived field is hidden from pickers but not deleted).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"custom_field_id": map[string]any{
					"type":        "integer",
					"description": "Custom field ID",
				},
				"archived": map[string]any{
					"type":        "boolean",
					"description": "true to archive, false to restore",
				},
			},
			"required": []string{"custom_field_id", "archived"},
		},
		Handler: Typed(r.setCustomFieldArchived),
	})

	r.register(&Tool{
		Name:        "list_project_custom_fields",
		Description: "List the custom fields attached to a project, with their project-scoped required/locked/default settings.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"query": map[string]any{
					"type":        "string",
					"description": "Filter by field name (optional)",
				},
				"page": map[string]any{
					"type":        "integer",
					"description": "Page number (0-based)",
					"default":     0,
				},
				"size": map[string]any{
					"type":        "integer",
					"description": "Items per page",
					"default":     10,
				},
			},
			"required": []string{"project_id"},
		},
		Handler: Typed(r.listProjectCustomFields),
	})

	r.register(&Tool{
		Name:        "get_project_custom_field",
		Description: "Get a single custom field's project-scoped settings (required, locked, default value) within a project.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"custom_field_id": map[string]any{
					"type":        "integer",
					"description": "Custom field ID",
				},
			},
			"required": []string{"project_id", "custom_field_id"},
		},
		Handler: Typed(r.getProjectCustomField),
	})

	r.register(&Tool{
		Name: "add_custom_fields_to_project",
		Description: "Attach one or more existing custom field definitions to a project, making them usable on that project's test cases. " +
			"Ids that don't exist are reported in not_attached (the API itself silently ignores them).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"custom_field_ids": map[string]any{
					"type":        "array",
					"description": "IDs of the custom fields to attach",
					"items":       map[string]any{"type": "integer"},
				},
			},
			"required": []string{"project_id", "custom_field_ids"},
		},
		Handler: Typed(r.addCustomFieldsToProject),
	})

	r.register(&Tool{
		Name: "remove_custom_field_from_project",
		Description: "Detach a custom field from a project. This does not delete the field definition or its values, only the project's use of it. " +
			"The API refuses while any test case in the project has a value for it (custom-field.in-use.test-case) — clear those first with bulk_remove_test_case_custom_fields.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"custom_field_id": map[string]any{
					"type":        "integer",
					"description": "Custom field ID",
				},
			},
			"required": []string{"project_id", "custom_field_id"},
		},
		Handler: Typed(r.removeCustomFieldFromProject),
	})

	r.register(&Tool{
		Name: "update_project_custom_field",
		Description: "Update a custom field's project-scoped settings: whether it's required, locked, or its " +
			"default value within this project. All fields are optional — only the ones you pass are changed.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"custom_field_id": map[string]any{
					"type":        "integer",
					"description": "Custom field ID",
				},
				"required": map[string]any{
					"type":        "boolean",
					"description": "Whether a value must be set on this project's test cases (optional)",
				},
				"locked": map[string]any{
					"type":        "boolean",
					"description": "Whether this project's use of the field is locked against further changes (optional)",
				},
				"default_custom_field_value_id": map[string]any{
					"type":        "integer",
					"description": "Custom field value ID to use as the default in this project (optional)",
				},
			},
			"required": []string{"project_id", "custom_field_id"},
		},
		Handler: Typed(r.updateProjectCustomField),
	})

	r.register(&Tool{
		Name: "create_custom_field_value",
		Description: "Create a new selectable value option for a custom field within a project (e.g. adding " +
			"\"Critical\" as a new Priority option). Use list_custom_field_values first to check the value " +
			"doesn't already exist.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"custom_field_id": map[string]any{
					"type":        "integer",
					"description": "Custom field ID to add the value to",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "The new value's display name",
				},
				"default": map[string]any{
					"type":        "boolean",
					"description": "Whether this becomes the field's default value in this project (optional)",
				},
			},
			"required": []string{"project_id", "custom_field_id", "name"},
		},
		Handler: Typed(r.createCustomFieldValue),
	})

	r.register(&Tool{
		Name: "update_custom_field_value",
		Description: "Rename a custom field value, or change its default/global flag. Renaming gives the value a NEW value_id and the old id stops " +
			"existing; test cases that had the value follow it to the new name (confirmed live). The API returns no body, so call " +
			"list_custom_field_values afterward to find the new id. global=true shares the value across all projects and cannot be undone. " +
			"Setting default=false clears the project default. All fields are optional — only the ones you pass are changed.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"value_id": map[string]any{
					"type":        "integer",
					"description": "Custom field value ID",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "New display name (optional)",
				},
				"default": map[string]any{
					"type":        "boolean",
					"description": "Whether this becomes the field's default value in this project (optional)",
				},
				"global": map[string]any{
					"type":        "boolean",
					"description": "Whether this value is shared globally rather than scoped to this project (optional)",
				},
			},
			"required": []string{"project_id", "value_id"},
		},
		Handler: Typed(r.updateCustomFieldValue),
	})

	r.register(&Tool{
		Name:        "delete_custom_field_value",
		Description: "Delete a custom field value option from a project. Test cases currently set to this value will lose it.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"project_id": map[string]any{
					"type":        "integer",
					"description": "Allure project ID",
				},
				"value_id": map[string]any{
					"type":        "integer",
					"description": "Custom field value ID",
				},
			},
			"required": []string{"project_id", "value_id"},
		},
		Handler: Typed(r.deleteCustomFieldValue),
	})
}

type createCustomFieldArgs struct {
	Name         string `json:"name"`
	Required     bool   `json:"required"`
	SingleSelect bool   `json:"single_select"`
}

func (r *Registry) createCustomField(ctx context.Context, args createCustomFieldArgs) (any, error) {
	if args.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	r.logger.Info("creating custom field", map[string]any{"name": args.Name})

	cf, err := r.allure.CreateCustomField(ctx, allure.CustomFieldCreateDto{
		Name:         args.Name,
		Required:     args.Required,
		SingleSelect: args.SingleSelect,
	})
	if err != nil {
		r.logger.Error("create custom field", err, map[string]any{"name": args.Name})
		return nil, fmt.Errorf("create custom field: %w", err)
	}

	return customFieldDtoToMap(*cf), nil
}

type getCustomFieldArgs struct {
	CustomFieldID int64 `json:"custom_field_id"`
}

func (r *Registry) getCustomField(ctx context.Context, args getCustomFieldArgs) (any, error) {
	if args.CustomFieldID == 0 {
		return nil, fmt.Errorf("custom_field_id is required (built-in fields like Epic/Feature/Story/Suite have negative ids)")
	}

	r.logger.Info("fetching custom field", map[string]any{"custom_field_id": args.CustomFieldID})

	cf, err := r.allure.GetCustomField(ctx, args.CustomFieldID)
	if err != nil {
		r.logger.Error("get custom field", err, map[string]any{"custom_field_id": args.CustomFieldID})
		return nil, fmt.Errorf("get custom field: %w", err)
	}

	return customFieldDtoToMap(*cf), nil
}

type updateCustomFieldArgs struct {
	CustomFieldID int64   `json:"custom_field_id"`
	Name          *string `json:"name"`
	Required      *bool   `json:"required"`
	SingleSelect  *bool   `json:"single_select"`
	Locked        *bool   `json:"locked"`
}

func (r *Registry) updateCustomField(ctx context.Context, args updateCustomFieldArgs) (any, error) {
	if args.CustomFieldID <= 0 {
		return nil, fmt.Errorf("custom_field_id must be positive — built-in fields (negative ids) can't be changed or removed")
	}
	if args.Name == nil && args.Required == nil && args.SingleSelect == nil && args.Locked == nil {
		return nil, fmt.Errorf("at least one field must be provided")
	}

	r.logger.Info("updating custom field", map[string]any{"custom_field_id": args.CustomFieldID})

	cf, err := r.allure.UpdateCustomField(ctx, args.CustomFieldID, allure.CustomFieldPatchDto{
		Name:         args.Name,
		Required:     args.Required,
		SingleSelect: args.SingleSelect,
		Locked:       args.Locked,
	})
	if err != nil {
		r.logger.Error("update custom field", err, map[string]any{"custom_field_id": args.CustomFieldID})
		return nil, fmt.Errorf("update custom field: %w", err)
	}

	return customFieldDtoToMap(*cf), nil
}

type deleteCustomFieldArgs struct {
	CustomFieldID int64 `json:"custom_field_id"`
}

func (r *Registry) deleteCustomField(ctx context.Context, args deleteCustomFieldArgs) (any, error) {
	if args.CustomFieldID <= 0 {
		return nil, fmt.Errorf("custom_field_id must be positive — built-in fields (negative ids) can't be changed or removed")
	}

	r.logger.Info("deleting custom field", map[string]any{"custom_field_id": args.CustomFieldID})

	if err := r.allure.DeleteCustomField(ctx, args.CustomFieldID); err != nil {
		r.logger.Error("delete custom field", err, map[string]any{"custom_field_id": args.CustomFieldID})
		return nil, fmt.Errorf("delete custom field: %w", err)
	}

	return map[string]any{"status": "deleted"}, nil
}

type setCustomFieldArchivedArgs struct {
	CustomFieldID int64 `json:"custom_field_id"`
	Archived      *bool `json:"archived"`
}

func (r *Registry) setCustomFieldArchived(ctx context.Context, args setCustomFieldArchivedArgs) (any, error) {
	if args.CustomFieldID <= 0 {
		return nil, fmt.Errorf("custom_field_id must be positive — built-in fields (negative ids) can't be changed or removed")
	}
	if args.Archived == nil {
		return nil, fmt.Errorf("archived must be specified (true or false)")
	}

	r.logger.Info("setting custom field archived state", map[string]any{
		"custom_field_id": args.CustomFieldID,
		"archived":        *args.Archived,
	})

	cf, err := r.allure.SetCustomFieldArchived(ctx, args.CustomFieldID, *args.Archived)
	if err != nil {
		r.logger.Error("set custom field archived", err, map[string]any{"custom_field_id": args.CustomFieldID})
		return nil, fmt.Errorf("set custom field archived: %w", err)
	}

	return customFieldDtoToMap(*cf), nil
}

type listProjectCustomFieldsArgs struct {
	ProjectID int64  `json:"project_id"`
	Query     string `json:"query"`
	Page      int    `json:"page"`
	Size      int    `json:"size"`
}

func (r *Registry) listProjectCustomFields(ctx context.Context, args listProjectCustomFieldsArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	size := args.Size
	if size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	r.logger.Info("listing project custom fields", map[string]any{"project_id": args.ProjectID})

	result, err := r.allure.ListProjectCustomFields(ctx, args.ProjectID, args.Query, args.Page, size)
	if err != nil {
		r.logger.Error("list project custom fields", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("list project custom fields: %w", err)
	}
	return result, nil
}

type getProjectCustomFieldArgs struct {
	ProjectID     int64 `json:"project_id"`
	CustomFieldID int64 `json:"custom_field_id"`
}

func (r *Registry) getProjectCustomField(ctx context.Context, args getProjectCustomFieldArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.CustomFieldID == 0 {
		return nil, fmt.Errorf("custom_field_id is required (built-in fields like Epic/Feature/Story/Suite have negative ids)")
	}

	r.logger.Info("fetching project custom field", map[string]any{
		"project_id":      args.ProjectID,
		"custom_field_id": args.CustomFieldID,
	})

	cfp, err := r.allure.GetProjectCustomField(ctx, args.ProjectID, args.CustomFieldID)
	if err != nil {
		r.logger.Error("get project custom field", err, map[string]any{
			"project_id":      args.ProjectID,
			"custom_field_id": args.CustomFieldID,
		})
		return nil, fmt.Errorf("get project custom field: %w", err)
	}

	m := map[string]any{
		"id":           cfp.ID,
		"project_id":   cfp.ProjectID,
		"name":         cfp.Name,
		"required":     cfp.Required,
		"locked":       cfp.Locked,
		"custom_field": customFieldDtoToMap(cfp.CustomField),
	}
	if cfp.DefaultCustomFieldValueID != nil {
		m["default_custom_field_value_id"] = *cfp.DefaultCustomFieldValueID
	}
	return m, nil
}

type addCustomFieldsToProjectArgs struct {
	ProjectID      int64   `json:"project_id"`
	CustomFieldIDs []int64 `json:"custom_field_ids"`
}

func (r *Registry) addCustomFieldsToProject(ctx context.Context, args addCustomFieldsToProjectArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if len(args.CustomFieldIDs) == 0 {
		return nil, fmt.Errorf("custom_field_ids must not be empty")
	}

	r.logger.Info("adding custom fields to project", map[string]any{
		"project_id":       args.ProjectID,
		"custom_field_ids": args.CustomFieldIDs,
	})

	if err := r.allure.AddCustomFieldsToProject(ctx, args.ProjectID, args.CustomFieldIDs); err != nil {
		r.logger.Error("add custom fields to project", err, map[string]any{"project_id": args.ProjectID})
		return nil, fmt.Errorf("add custom fields to project: %w", err)
	}

	// The API answers success even for ids that don't exist and attaches
	// nothing (confirmed live), so check each one.
	var missing []int64
	for _, id := range args.CustomFieldIDs {
		if _, err := r.allure.GetProjectCustomField(ctx, args.ProjectID, id); errors.Is(err, allure.ErrCustomFieldNotAttached) {
			missing = append(missing, id)
		}
	}
	if len(missing) == len(args.CustomFieldIDs) {
		return nil, fmt.Errorf("none of the custom fields were attached — check the ids exist (get_custom_field): %v", missing)
	}
	result := map[string]any{"status": "added", "count": len(args.CustomFieldIDs) - len(missing)}
	if len(missing) > 0 {
		result["not_attached"] = missing
	}
	return result, nil
}

type removeCustomFieldFromProjectArgs struct {
	ProjectID     int64 `json:"project_id"`
	CustomFieldID int64 `json:"custom_field_id"`
}

func (r *Registry) removeCustomFieldFromProject(ctx context.Context, args removeCustomFieldFromProjectArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.CustomFieldID <= 0 {
		return nil, fmt.Errorf("custom_field_id must be positive — built-in fields (negative ids) can't be changed or removed")
	}

	r.logger.Info("removing custom field from project", map[string]any{
		"project_id":      args.ProjectID,
		"custom_field_id": args.CustomFieldID,
	})

	if err := r.allure.RemoveCustomFieldFromProject(ctx, args.ProjectID, args.CustomFieldID); err != nil {
		r.logger.Error("remove custom field from project", err, map[string]any{
			"project_id":      args.ProjectID,
			"custom_field_id": args.CustomFieldID,
		})
		return nil, fmt.Errorf("remove custom field from project: %w", err)
	}

	return map[string]any{"status": "removed"}, nil
}

type updateProjectCustomFieldArgs struct {
	ProjectID                 int64  `json:"project_id"`
	CustomFieldID             int64  `json:"custom_field_id"`
	Required                  *bool  `json:"required"`
	Locked                    *bool  `json:"locked"`
	DefaultCustomFieldValueID *int64 `json:"default_custom_field_value_id"`
}

func (r *Registry) updateProjectCustomField(ctx context.Context, args updateProjectCustomFieldArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.CustomFieldID == 0 {
		return nil, fmt.Errorf("custom_field_id is required (built-in fields like Epic/Feature/Story/Suite have negative ids)")
	}
	if args.Required == nil && args.Locked == nil && args.DefaultCustomFieldValueID == nil {
		return nil, fmt.Errorf("at least one field must be provided")
	}

	r.logger.Info("updating project custom field", map[string]any{
		"project_id":      args.ProjectID,
		"custom_field_id": args.CustomFieldID,
	})

	req := allure.CustomFieldProjectPatchDto{
		Required:                  args.Required,
		Locked:                    args.Locked,
		DefaultCustomFieldValueID: args.DefaultCustomFieldValueID,
	}
	if err := r.allure.UpdateProjectCustomField(ctx, args.ProjectID, args.CustomFieldID, req); err != nil {
		r.logger.Error("update project custom field", err, map[string]any{
			"project_id":      args.ProjectID,
			"custom_field_id": args.CustomFieldID,
		})
		return nil, fmt.Errorf("update project custom field: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

type createCustomFieldValueArgs struct {
	ProjectID     int64  `json:"project_id"`
	CustomFieldID int64  `json:"custom_field_id"`
	Name          string `json:"name"`
	Default       bool   `json:"default"`
}

func (r *Registry) createCustomFieldValue(ctx context.Context, args createCustomFieldValueArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.CustomFieldID == 0 {
		return nil, fmt.Errorf("custom_field_id is required (built-in fields like Epic/Feature/Story/Suite have negative ids)")
	}
	if args.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	r.logger.Info("creating custom field value", map[string]any{
		"project_id":      args.ProjectID,
		"custom_field_id": args.CustomFieldID,
		"name":            args.Name,
	})

	cfv, err := r.allure.CreateCustomFieldValue(ctx, args.ProjectID, allure.CustomFieldValueProjectCreateDto{
		CustomField: allure.IDOnlyDto{ID: args.CustomFieldID},
		Name:        args.Name,
		Default:     args.Default,
	})
	if err != nil {
		r.logger.Error("create custom field value", err, map[string]any{
			"project_id":      args.ProjectID,
			"custom_field_id": args.CustomFieldID,
		})
		return nil, fmt.Errorf("create custom field value: %w", err)
	}

	return map[string]any{
		"id":           cfv.ID,
		"name":         cfv.Name,
		"global":       cfv.Global,
		"custom_field": customFieldDtoToMap(cfv.CustomField),
	}, nil
}

type updateCustomFieldValueArgs struct {
	ProjectID int64   `json:"project_id"`
	ValueID   int64   `json:"value_id"`
	Name      *string `json:"name"`
	Default   *bool   `json:"default"`
	Global    *bool   `json:"global"`
}

func (r *Registry) updateCustomFieldValue(ctx context.Context, args updateCustomFieldValueArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.ValueID <= 0 {
		return nil, fmt.Errorf("value_id must be positive")
	}
	if args.Name == nil && args.Default == nil && args.Global == nil {
		return nil, fmt.Errorf("at least one field must be provided")
	}

	r.logger.Info("updating custom field value", map[string]any{
		"project_id": args.ProjectID,
		"value_id":   args.ValueID,
	})

	req := allure.CustomFieldValueProjectPatchDto{
		Name:    args.Name,
		Default: args.Default,
		Global:  args.Global,
	}
	if err := r.allure.UpdateCustomFieldValue(ctx, args.ProjectID, args.ValueID, req); err != nil {
		r.logger.Error("update custom field value", err, map[string]any{
			"project_id": args.ProjectID,
			"value_id":   args.ValueID,
		})
		return nil, fmt.Errorf("update custom field value: %w", err)
	}

	return map[string]any{"status": "updated"}, nil
}

type deleteCustomFieldValueArgs struct {
	ProjectID int64 `json:"project_id"`
	ValueID   int64 `json:"value_id"`
}

func (r *Registry) deleteCustomFieldValue(ctx context.Context, args deleteCustomFieldValueArgs) (any, error) {
	if args.ProjectID <= 0 {
		return nil, fmt.Errorf("project_id must be positive")
	}
	if args.ValueID <= 0 {
		return nil, fmt.Errorf("value_id must be positive")
	}

	r.logger.Info("deleting custom field value", map[string]any{
		"project_id": args.ProjectID,
		"value_id":   args.ValueID,
	})

	if err := r.allure.DeleteCustomFieldValue(ctx, args.ProjectID, args.ValueID); err != nil {
		r.logger.Error("delete custom field value", err, map[string]any{
			"project_id": args.ProjectID,
			"value_id":   args.ValueID,
		})
		return nil, fmt.Errorf("delete custom field value: %w", err)
	}

	return map[string]any{"status": "deleted"}, nil
}
