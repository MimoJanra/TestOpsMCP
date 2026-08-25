package tools

import (
	"context"
	"net/http"
	"testing"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
)

const pagedEmptyJSON = `{"content":[],"last":true,"number":0,"size":20,"totalElements":0}`

func TestGetTestCaseTags_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `[{"id":1,"name":"smoke"}]`))
	res, err := r.getTestCaseTags(context.Background(), getTestCaseTagsArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags := res.(map[string]any)["tags"].([]map[string]any)
	if len(tags) != 1 || tags[0]["name"] != "smoke" {
		t.Errorf("unexpected tags: %+v", tags)
	}
}

func TestGetTestCaseTags_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.getTestCaseTags(context.Background(), getTestCaseTagsArgs{}); err == nil {
		t.Fatal("expected error for non-positive test_case_id")
	}
}

func TestSetTestCaseTags_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.setTestCaseTags(context.Background(), setTestCaseTagsArgs{TestCaseID: 1, Tags: []allure.TestTagDto{{ID: 1, Name: "smoke"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["count"] != 1 {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestCreateTestTag_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"id":42,"name":"regression"}`))
	res, err := r.createTestTag(context.Background(), createTestTagArgs{Name: "regression"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := res.(map[string]any)
	if m["id"] != int64(42) || m["name"] != "regression" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestCreateTestTag_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.createTestTag(context.Background(), createTestTagArgs{}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestGetTestCaseIssues_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `[{"id":1,"displayName":"PROJ-1","url":"http://x","integrationId":2,"closed":true}]`))
	res, err := r.getTestCaseIssues(context.Background(), getTestCaseIssuesArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	issues := res.(map[string]any)["issues"].([]map[string]any)
	if len(issues) != 1 || issues[0]["display_name"] != "PROJ-1" {
		t.Errorf("unexpected issues: %+v", issues)
	}
}

func TestSetTestCaseIssues_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.setTestCaseIssues(context.Background(), setTestCaseIssuesArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "updated" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestGetTestCaseExamples_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `[[{"name":"a","value":"1"}]]`))
	res, err := r.getTestCaseExamples(context.Background(), getTestCaseExamplesArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["count"] != 1 {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestSetTestCaseExamples_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	args := setTestCaseExamplesArgs{TestCaseID: 1, Examples: [][]allure.TestCaseExampleParam{{{Name: "a", Value: "1"}}}}
	res, err := r.setTestCaseExamples(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["rows"] != 1 {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestListTestCaseVersions_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `[{"id":1,"title":"v1"}]`))
	res, err := r.listTestCaseVersions(context.Background(), listTestCaseVersionsArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	versions := res.(map[string]any)["versions"].([]map[string]any)
	if len(versions) != 1 || versions[0]["title"] != "v1" {
		t.Errorf("unexpected versions: %+v", versions)
	}
}

func TestCreateTestCaseVersion_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"id":5,"title":"v1","description":"d"}`))
	res, err := r.createTestCaseVersion(context.Background(), createTestCaseVersionArgs{TestCaseID: 1, Title: "v1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["id"] != int64(5) {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestCreateTestCaseVersion_RequiresTitle(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.createTestCaseVersion(context.Background(), createTestCaseVersionArgs{TestCaseID: 1}); err == nil {
		t.Fatal("expected error for missing title")
	}
}

func TestRestoreTestCaseVersion_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.restoreTestCaseVersion(context.Background(), restoreTestCaseVersionArgs{VersionID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "restored" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestGetTestCaseAttachments_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{
		"content":[{"id":1,"name":"a.png","contentType":"image/png","size":100,"createdDate":123}],
		"last":true,"number":0,"size":20,"totalElements":1
	}`))
	res, err := r.getTestCaseAttachments(context.Background(), getTestCaseAttachmentsArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["total"] != 1 {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestDeleteTestCaseAttachment_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.deleteTestCaseAttachment(context.Background(), deleteTestCaseAttachmentArgs{AttachmentID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "deleted" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestSearchTestCases_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, pagedEmptyJSON))
	res, err := r.searchTestCases(context.Background(), searchTestCasesArgs{ProjectID: 1, Query: "name ~ 'x'"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["total"] != 0 {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestSearchTestCases_RequiresQuery(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.searchTestCases(context.Background(), searchTestCasesArgs{ProjectID: 1}); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestListDeletedTestCases_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, pagedEmptyJSON))
	res, err := r.listDeletedTestCases(context.Background(), listDeletedTestCasesArgs{ProjectID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["is_last"] != true {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestGetTestCaseScenario_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"steps":[]}`))
	res, err := r.getTestCaseScenario(context.Background(), getTestCaseScenarioArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestGetTestCaseSteps_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"steps":[]}`))
	res, err := r.getTestCaseSteps(context.Background(), getTestCaseStepsArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestDeleteTestCaseScenario_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.deleteTestCaseScenario(context.Background(), deleteTestCaseScenarioArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "deleted" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestMoveTestCaseStep_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.moveTestCaseStep(context.Background(), moveTestCaseStepArgs{StepID: 1, AfterID: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "moved" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestMoveTestCaseStep_ValidatesInput(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.moveTestCaseStep(context.Background(), moveTestCaseStepArgs{}); err == nil {
		t.Fatal("expected error for non-positive step_id")
	}
}

func TestCopyTestCaseStep_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.copyTestCaseStep(context.Background(), copyTestCaseStepArgs{StepID: 1, BeforeID: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "copied" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestGetTestCaseRelations_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `[{"id":1,"target":{"id":2,"name":"other"}}]`))
	res, err := r.getTestCaseRelations(context.Background(), getTestCaseRelationsArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	relations := res.(map[string]any)["relations"].([]map[string]any)
	if len(relations) != 1 || relations[0]["target_name"] != "other" {
		t.Errorf("unexpected relations: %+v", relations)
	}
}

func TestSetTestCaseRelations_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	args := setTestCaseRelationsArgs{TestCaseID: 1}
	args.Relations = append(args.Relations, struct {
		TargetID   int64  `json:"target_id"`
		TargetName string `json:"target_name"`
	}{TargetID: 2, TargetName: "other"})

	res, err := r.setTestCaseRelations(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["count"] != 1 {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestListMutedTestCases_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, pagedEmptyJSON))
	res, err := r.listMutedTestCases(context.Background(), listMutedTestCasesArgs{ProjectID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["total"] != 0 {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestGetTestCaseAudit_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"content":[{"field":"name"}],"last":true,"number":0,"size":20,"totalElements":1}`))
	res, err := r.getTestCaseAudit(context.Background(), getTestCaseAuditArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["total"] != 1 {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestValidateTestCaseQuery_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"valid":true}`))
	res, err := r.validateTestCaseQuery(context.Background(), validateTestCaseQueryArgs{ProjectID: 1, Query: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestValidateTestCaseQuery_RequiresQuery(t *testing.T) {
	r := newTestRegistry(t)
	if _, err := r.validateTestCaseQuery(context.Background(), validateTestCaseQueryArgs{ProjectID: 1}); err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestSuggestTestCases_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"suggestions":[]}`))
	res, err := r.suggestTestCases(context.Background(), suggestTestCasesArgs{ProjectID: 1, Query: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestGetTestCaseWorkflow_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"states":[]}`))
	res, err := r.getTestCaseWorkflow(context.Background(), getTestCaseWorkflowArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestGetTestCaseKeys_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `[{"id":1,"integrationId":2,"name":"JIRA-1","url":"http://x"}]`))
	res, err := r.getTestCaseKeys(context.Background(), getTestCaseKeysArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	keys := res.(map[string]any)["keys"].([]map[string]any)
	if len(keys) != 1 || keys[0]["name"] != "JIRA-1" {
		t.Errorf("unexpected keys: %+v", keys)
	}
}

func TestSetTestCaseKeys_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.setTestCaseKeys(context.Background(), setTestCaseKeysArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "updated" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestGetTestCaseScenarioFromRun_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"steps":[]}`))
	res, err := r.getTestCaseScenarioFromRun(context.Background(), getTestCaseScenarioFromRunArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestDetachTestCaseAutomation_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.detachTestCaseAutomation(context.Background(), detachTestCaseAutomationArgs{TestCaseID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "detached" {
		t.Errorf("unexpected result: %v", res)
	}
}

func TestGetTestCaseVersionData_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{"id":1}`))
	res, err := r.getTestCaseVersionData(context.Background(), getTestCaseVersionDataArgs{VersionID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestDeleteTestCaseVersion_Handler(t *testing.T) {
	r := newTestRegistryWithServer(t, jsonHandler(http.StatusOK, `{}`))
	res, err := r.deleteTestCaseVersion(context.Background(), deleteTestCaseVersionArgs{VersionID: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.(map[string]any)["status"] != "deleted" {
		t.Errorf("unexpected result: %v", res)
	}
}
