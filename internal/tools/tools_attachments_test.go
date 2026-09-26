package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeAttachServer mirrors the live shapes on project 408: test case 19575 has
// step 77076 whose expected-result container is 77081, and step 77077 with no
// expected result; an upload returns one attachment row, and attaching it to a
// step is a child node created with parentId + attachmentId.
type fakeAttachServer struct {
	uploads     int
	fileName    string
	contentType string
	content     string
	stepBody    map[string]any
}

func (f *fakeAttachServer) mux(t *testing.T) *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("/api/testcase/19575/step", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"root": map[string]any{"children": []int{77076, 77077}},
			"scenarioSteps": map[string]any{
				"77076": map[string]any{"id": 77076, "body": "a", "expectedResultId": 77081},
				"77077": map[string]any{"id": 77077, "body": "b"},
			},
		})
	})
	m.HandleFunc("/api/testcase/attachment", func(w http.ResponseWriter, r *http.Request) {
		f.uploads++
		if r.URL.Query().Get("testCaseId") != "19575" {
			t.Errorf("upload testCaseId = %q, want 19575", r.URL.Query().Get("testCaseId"))
		}
		mt, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mt != "multipart/form-data" {
			t.Fatalf("upload must be multipart/form-data, got %q", r.Header.Get("Content-Type"))
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		part, err := mr.NextPart()
		if err != nil {
			t.Fatalf("read part: %v", err)
		}
		if part.FormName() != "file" {
			t.Errorf("form field = %q, want file", part.FormName())
		}
		f.fileName = part.FileName()
		f.contentType = part.Header.Get("Content-Type")
		b, _ := io.ReadAll(part)
		f.content = string(b)
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 7160, "name": f.fileName, "contentType": f.contentType, "contentLength": len(b), "entity": "test_case"}})
	})
	m.HandleFunc("/api/testcase/step", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&f.stepBody)
		_ = json.NewEncoder(w).Encode(map[string]any{"createdStepId": 77090})
	})
	return m
}

func newAttachTestRegistry(t *testing.T) (*Registry, *fakeAttachServer) {
	f := &fakeAttachServer{}
	return newRelationsTestRegistry(t, f.mux(t)), f
}

func TestUploadTestCaseAttachment_Base64ToStep(t *testing.T) {
	r, f := newAttachTestRegistry(t)
	res, err := r.uploadTestCaseAttachment(context.Background(), uploadTestCaseAttachmentArgs{
		TestCaseID: 19575, ContentBase64: base64.StdEncoding.EncodeToString([]byte("hello")), FileName: "notes.txt", StepID: 77077,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := res.(map[string]any)
	if out["attachment_id"] != int64(7160) || out["node_id"] != int64(77090) {
		t.Errorf("unexpected result: %+v", out)
	}
	if f.fileName != "notes.txt" || f.content != "hello" || !strings.HasPrefix(f.contentType, "text/plain") {
		t.Errorf("uploaded %q (%s) = %q", f.fileName, f.contentType, f.content)
	}
	if f.stepBody["parentId"] != float64(77077) || f.stepBody["attachmentId"] != float64(7160) || f.stepBody["testCaseId"] != float64(19575) {
		t.Errorf("attach node body = %+v, want parentId 77077, attachmentId 7160", f.stepBody)
	}
}

func TestUploadTestCaseAttachment_ExpectedResultTarget(t *testing.T) {
	r, f := newAttachTestRegistry(t)
	_, err := r.uploadTestCaseAttachment(context.Background(), uploadTestCaseAttachmentArgs{
		TestCaseID: 19575, ContentBase64: base64.StdEncoding.EncodeToString([]byte("x")), FileName: "a.png", StepID: 77076, Target: "expected_result",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.stepBody["parentId"] != float64(77081) {
		t.Errorf("parentId = %v, want the expected-result container 77081", f.stepBody["parentId"])
	}
	if f.contentType != "image/png" {
		t.Errorf("content type = %q, want image/png from the extension", f.contentType)
	}

	// A step without an expected result: refuse before uploading anything.
	f.uploads = 0
	if _, err := r.uploadTestCaseAttachment(context.Background(), uploadTestCaseAttachmentArgs{
		TestCaseID: 19575, ContentBase64: base64.StdEncoding.EncodeToString([]byte("x")), FileName: "a.png", StepID: 77077, Target: "expected_result",
	}); err == nil {
		t.Error("expected error for a step with no expected result")
	}
	if f.uploads != 0 {
		t.Error("nothing should be uploaded when the target step is invalid")
	}
}

func TestUploadTestCaseAttachment_FilePathOnlyWhenLocal(t *testing.T) {
	r, f := newAttachTestRegistry(t)
	path := filepath.Join(t.TempDir(), "report.csv")
	if err := os.WriteFile(path, []byte("a,b"), 0o600); err != nil {
		t.Fatal(err)
	}
	args := uploadTestCaseAttachmentArgs{TestCaseID: 19575, FilePath: path}

	if _, err := r.uploadTestCaseAttachment(context.Background(), args); err == nil {
		t.Fatal("file_path must be refused unless local file access was enabled at startup (HTTP mode)")
	}
	if f.uploads != 0 {
		t.Error("refused file_path must not upload")
	}

	r.AllowLocalFiles()
	res, err := r.uploadTestCaseAttachment(context.Background(), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.fileName != "report.csv" || f.content != "a,b" {
		t.Errorf("uploaded %q = %q", f.fileName, f.content)
	}
	if _, attached := res.(map[string]any)["node_id"]; attached {
		t.Error("without step_id the file should only be attached to the test case")
	}
}

func TestUploadTestCaseAttachment_Validation(t *testing.T) {
	r, f := newAttachTestRegistry(t)
	for name, args := range map[string]uploadTestCaseAttachmentArgs{
		"no test case": {ContentBase64: "eA==", FileName: "a"},
		"no content":   {TestCaseID: 19575},
		"both sources": {TestCaseID: 19575, ContentBase64: "eA==", FileName: "a", FilePath: "x"},
		"no file name": {TestCaseID: 19575, ContentBase64: "eA=="},
		"bad base64":   {TestCaseID: 19575, ContentBase64: "not base64!", FileName: "a"},
		"bad target":   {TestCaseID: 19575, ContentBase64: "eA==", FileName: "a", Target: "body"},
		"missing step": {TestCaseID: 19575, ContentBase64: "eA==", FileName: "a", StepID: 1},
	} {
		if _, err := r.uploadTestCaseAttachment(context.Background(), args); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
	if f.uploads != 0 {
		t.Errorf("invalid calls must not upload, got %d uploads", f.uploads)
	}
}

// TestAddTestCaseStepTable guards the live-observed shape of a step table: a
// CSV attachment (text/csv, no trailing newline) shown in the step as a child
// node — not a table node inside bodyJson.
func TestAddTestCaseStepTable(t *testing.T) {
	r, f := newAttachTestRegistry(t)
	res, err := r.addTestCaseStepTable(context.Background(), addTestCaseStepTableArgs{
		TestCaseID: 19575, StepID: 77077, Rows: [][]string{{"1", "6", "2"}, {"a,b", "5", "3"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "1,6,2\n\"a,b\",5,3"; f.content != want {
		t.Errorf("csv = %q, want %q", f.content, want)
	}
	if f.contentType != "text/csv" || f.fileName != "Table" {
		t.Errorf("uploaded %q as %q, want \"Table\" as text/csv", f.fileName, f.contentType)
	}
	if f.stepBody["attachmentId"] != float64(7160) || f.stepBody["parentId"] != float64(77077) {
		t.Errorf("attach node body = %+v", f.stepBody)
	}
	if res.(map[string]any)["node_id"] != int64(77090) {
		t.Errorf("unexpected result: %+v", res)
	}

	if _, err := r.addTestCaseStepTable(context.Background(), addTestCaseStepTableArgs{TestCaseID: 19575, StepID: 77077}); err == nil {
		t.Error("expected error for empty rows")
	}
}
