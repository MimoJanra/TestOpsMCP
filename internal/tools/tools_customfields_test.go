package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCreateCustomField_Handler(t *testing.T) {
	var gotBody map[string]any
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/api/cf" || req.Method != http.MethodPost {
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL.Path)
		}
		_ = json.NewDecoder(req.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":5,"name":"Severity","required":true}`))
	})

	res, err := r.createCustomField(context.Background(), createCustomFieldArgs{Name: "Severity", Required: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := res.(map[string]any)
	if m["id"] != int64(5) || m["name"] != "Severity" {
		t.Errorf("unexpected result: %v", m)
	}
	if gotBody["name"] != "Severity" || gotBody["required"] != true {
		t.Errorf("unexpected request body: %v", gotBody)
	}
}

func TestCreateCustomField_RequiresName(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.createCustomField(context.Background(), createCustomFieldArgs{}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGetCustomField_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"id":5,"name":"Severity","archived":false}`))
	res, err := r.getCustomField(context.Background(), getCustomFieldArgs{CustomFieldID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["name"] != "Severity" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestGetCustomField_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.getCustomField(context.Background(), getCustomFieldArgs{}); err == nil {
		t.Fatal("expected error for non-positive custom_field_id")
	}
}

func TestUpdateCustomField_Handler(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotPath = req.URL.Path
		_ = json.NewDecoder(req.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":5,"name":"Sev2"}`))
	})
	name := "Sev2"
	res, err := r.updateCustomField(context.Background(), updateCustomFieldArgs{CustomFieldID: 5, Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/cf/5" {
		t.Errorf("path = %q, want /api/cf/5", gotPath)
	}
	if gotBody["name"] != "Sev2" {
		t.Errorf("unexpected request body: %v", gotBody)
	}
	if res.(map[string]any)["name"] != "Sev2" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestUpdateCustomField_RequiresAtLeastOneField(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.updateCustomField(context.Background(), updateCustomFieldArgs{CustomFieldID: 5}); err == nil {
		t.Fatal("expected error when no fields are provided")
	}
}

func TestDeleteCustomField_Handler(t *testing.T) {
	var gotPath, gotMethod string
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotPath, gotMethod = req.URL.Path, req.Method
		w.WriteHeader(http.StatusNoContent)
	})
	res, err := r.deleteCustomField(context.Background(), deleteCustomFieldArgs{CustomFieldID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/cf/5" || gotMethod != http.MethodDelete {
		t.Errorf("request = %s %s, want DELETE /api/cf/5", gotMethod, gotPath)
	}
	if res.(map[string]any)["status"] != "deleted" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestSetCustomFieldArchived_Handler(t *testing.T) {
	var gotURL string
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotURL = req.URL.String()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":5,"archived":true}`))
	})
	archived := true
	res, err := r.setCustomFieldArchived(context.Background(), setCustomFieldArchivedArgs{CustomFieldID: 5, Archived: &archived})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotURL != "/api/cf/5/archived?archived=true" {
		t.Errorf("url = %q, want /api/cf/5/archived?archived=true", gotURL)
	}
	if res.(map[string]any)["archived"] != true {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestSetCustomFieldArchived_RequiresArchivedValue(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.setCustomFieldArchived(context.Background(), setCustomFieldArchivedArgs{CustomFieldID: 5}); err == nil {
		t.Fatal("expected error when archived is not specified")
	}
}

func TestListProjectCustomFields_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"content":[{"id":1,"name":"Priority"}],"totalElements":1}`))
	res, err := r.listProjectCustomFields(context.Background(), listProjectCustomFieldsArgs{ProjectID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := res.(map[string]any)
	if m["totalElements"] != float64(1) {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestListProjectCustomFields_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.listProjectCustomFields(context.Background(), listProjectCustomFieldsArgs{}); err == nil {
		t.Fatal("expected error for non-positive project_id")
	}
}

func TestGetProjectCustomField_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"id":9,"projectId":1,"required":true,"customField":{"id":5,"name":"Priority"}}`))
	res, err := r.getProjectCustomField(context.Background(), getProjectCustomFieldArgs{ProjectID: 1, CustomFieldID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := res.(map[string]any)
	if m["required"] != true {
		t.Errorf("unexpected result: %v", m)
	}
	cf, ok := m["custom_field"].(map[string]any)
	if !ok || cf["name"] != "Priority" {
		t.Errorf("unexpected nested custom_field: %v", m["custom_field"])
	}
}

func TestGetProjectCustomField_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.getProjectCustomField(context.Background(), getProjectCustomFieldArgs{ProjectID: 1}); err == nil {
		t.Fatal("expected error for non-positive custom_field_id")
	}
}

func TestAddCustomFieldsToProject_Handler(t *testing.T) {
	var gotURL string
	var gotBody map[string]any
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			// Attachment check: field 6 attached, field 5 not (empty body).
			if req.URL.Query().Get("customFieldId") == "6" {
				_, _ = w.Write([]byte(`{"id":1}`))
			}
			return
		}
		gotURL = req.URL.String()
		_ = json.NewDecoder(req.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusAccepted)
	})
	res, err := r.addCustomFieldsToProject(context.Background(), addCustomFieldsToProjectArgs{ProjectID: 1, CustomFieldIDs: []int64{5, 6}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nb := res.(map[string]any)["not_attached"]; fmt.Sprint(nb) != "[5]" {
		t.Errorf("not_attached = %v, want [5]", nb)
	}
	if gotURL != "/api/cfproject/add-to-project?projectId=1" {
		t.Errorf("url = %q, want /api/cfproject/add-to-project?projectId=1", gotURL)
	}
	ids, _ := gotBody["ids"].([]any)
	if len(ids) != 2 {
		t.Errorf("request body ids = %v, want 2 entries", gotBody["ids"])
	}
	if res.(map[string]any)["count"] != 1 {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestAddCustomFieldsToProject_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.addCustomFieldsToProject(context.Background(), addCustomFieldsToProjectArgs{ProjectID: 1}); err == nil {
		t.Fatal("expected error for empty custom_field_ids")
	}
}

func TestRemoveCustomFieldFromProject_Handler(t *testing.T) {
	var gotURL, gotMethod string
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotURL, gotMethod = req.URL.String(), req.Method
		w.WriteHeader(http.StatusNoContent)
	})
	res, err := r.removeCustomFieldFromProject(context.Background(), removeCustomFieldFromProjectArgs{ProjectID: 1, CustomFieldID: 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete || gotURL != "/api/cfproject/remove?customFieldId=5&projectId=1" {
		t.Errorf("request = %s %s, want DELETE /api/cfproject/remove?customFieldId=5&projectId=1", gotMethod, gotURL)
	}
	if res.(map[string]any)["status"] != "removed" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestUpdateProjectCustomField_Handler(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotPath = req.URL.Path
		_ = json.NewDecoder(req.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	})
	required := true
	res, err := r.updateProjectCustomField(context.Background(), updateProjectCustomFieldArgs{
		ProjectID: 1, CustomFieldID: 5, Required: &required,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/project/1/cf/5" {
		t.Errorf("path = %q, want /api/project/1/cf/5", gotPath)
	}
	if gotBody["required"] != true {
		t.Errorf("unexpected request body: %v", gotBody)
	}
	if res.(map[string]any)["status"] != "updated" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestUpdateProjectCustomField_RequiresAtLeastOneField(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.updateProjectCustomField(context.Background(), updateProjectCustomFieldArgs{ProjectID: 1, CustomFieldID: 5}); err == nil {
		t.Fatal("expected error when no fields are provided")
	}
}

func TestCreateCustomFieldValue_Handler(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotPath = req.URL.Path
		_ = json.NewDecoder(req.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":100,"name":"Critical","customField":{"id":5,"name":"Priority"}}`))
	})
	res, err := r.createCustomFieldValue(context.Background(), createCustomFieldValueArgs{
		ProjectID: 1, CustomFieldID: 5, Name: "Critical",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/project/1/cfv" {
		t.Errorf("path = %q, want /api/project/1/cfv", gotPath)
	}
	cf, _ := gotBody["customField"].(map[string]any)
	if cf["id"] != float64(5) || gotBody["name"] != "Critical" {
		t.Errorf("unexpected request body: %v", gotBody)
	}
	if res.(map[string]any)["id"] != int64(100) {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestCreateCustomFieldValue_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.createCustomFieldValue(context.Background(), createCustomFieldValueArgs{ProjectID: 1, CustomFieldID: 5}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUpdateCustomFieldValue_Handler(t *testing.T) {
	var gotPath string
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotPath = req.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})
	name := "Blocker"
	res, err := r.updateCustomFieldValue(context.Background(), updateCustomFieldValueArgs{ProjectID: 1, ValueID: 100, Name: &name})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/project/1/cfv/100" {
		t.Errorf("path = %q, want /api/project/1/cfv/100", gotPath)
	}
	if res.(map[string]any)["status"] != "updated" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestUpdateCustomFieldValue_RequiresAtLeastOneField(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.updateCustomFieldValue(context.Background(), updateCustomFieldValueArgs{ProjectID: 1, ValueID: 100}); err == nil {
		t.Fatal("expected error when no fields are provided")
	}
}

func TestDeleteCustomFieldValue_Handler(t *testing.T) {
	var gotPath, gotMethod string
	r := newTestRegistryWithServer(t, func(w http.ResponseWriter, req *http.Request) {
		gotPath, gotMethod = req.URL.Path, req.Method
		w.WriteHeader(http.StatusNoContent)
	})
	res, err := r.deleteCustomFieldValue(context.Background(), deleteCustomFieldValueArgs{ProjectID: 1, ValueID: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/api/project/1/cfv/100" || gotMethod != http.MethodDelete {
		t.Errorf("request = %s %s, want DELETE /api/project/1/cfv/100", gotMethod, gotPath)
	}
	if res.(map[string]any)["status"] != "deleted" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestDeleteCustomFieldValue_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.deleteCustomFieldValue(context.Background(), deleteCustomFieldValueArgs{ProjectID: 1}); err == nil {
		t.Fatal("expected error for non-positive value_id")
	}
}
