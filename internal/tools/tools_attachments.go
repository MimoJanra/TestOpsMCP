package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/MimoJanra/TestOpsMCP/internal/adapters/allure"
)

// maxAttachmentBytes caps a single upload so one call can't pull an arbitrarily
// large file into memory (and into an MCP message, for content_base64).
const maxAttachmentBytes = 20 << 20

var attachTargetSchema = map[string]any{
	"type":        "string",
	"enum":        []string{"step", "expected_result"},
	"description": "Where to show it within the step: the step itself (default) or the step's Expected Result",
	"default":     "step",
}

func (r *Registry) registerAttachmentTools() {
	r.register(&Tool{
		Name: "upload_test_case_attachment",
		Description: "Upload a file to a test case and optionally show it in a step (like the step's \"File\" block in " +
			"the web UI). Pass the file as file_path (local stdio mode only — refused on a shared HTTP server, where it " +
			"would read the server's disk) or as content_base64 + file_name. Without step_id the file is only attached " +
			"to the test case. Use add_test_case_step_table for tables.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id":   map[string]any{"type": "integer", "description": "Allure test case ID"},
				"file_path":      map[string]any{"type": "string", "description": "Path to a local file (stdio mode only)"},
				"content_base64": map[string]any{"type": "string", "description": "File content, base64-encoded (alternative to file_path)"},
				"file_name":      map[string]any{"type": "string", "description": "File name to show (required with content_base64; defaults to the file_path base name)"},
				"content_type":   map[string]any{"type": "string", "description": "MIME type (optional — guessed from the file name or content)"},
				"step_id":        map[string]any{"type": "integer", "description": "Step to show the file in (optional)"},
				"target":         attachTargetSchema,
			},
			"required": []string{"test_case_id"},
		},
		Handler: Typed(r.uploadTestCaseAttachment),
	})

	r.register(&Tool{
		Name: "get_test_case_attachment_content",
		Description: "Download a test case attachment's content (find ids with get_test_case_attachments or get_test_case_steps). " +
			"Text content (text/*, JSON, CSV, XML) is returned as text, anything else as content_base64. In local stdio mode " +
			"pass save_to to write the file to disk instead.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"attachment_id": map[string]any{"type": "integer", "description": "Attachment ID"},
				"save_to":       map[string]any{"type": "string", "description": "Local file path to save the content to (stdio mode only; the file must not exist yet)"},
			},
			"required": []string{"attachment_id"},
		},
		Handler: Typed(r.getTestCaseAttachmentContent),
	})

	r.register(&Tool{
		Name: "add_test_case_step_table",
		Description: "Add a table to a test case step (like the step's \"Table\" block in the web UI). In Allure a step " +
			"table is a CSV attachment shown in the step, so this uploads the rows as CSV and attaches it.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"test_case_id": map[string]any{"type": "integer", "description": "Allure test case ID"},
				"step_id":      map[string]any{"type": "integer", "description": "Step to add the table to"},
				"rows": map[string]any{
					"type":        "array",
					"description": "Table rows, each an array of cell strings",
					"items":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"name":   map[string]any{"type": "string", "description": "Table name (default \"Table\")"},
				"target": attachTargetSchema,
			},
			"required": []string{"test_case_id", "step_id", "rows"},
		},
		Handler: Typed(r.addTestCaseStepTable),
	})
}

type uploadTestCaseAttachmentArgs struct {
	TestCaseID    int64  `json:"test_case_id"`
	FilePath      string `json:"file_path"`
	ContentBase64 string `json:"content_base64"`
	FileName      string `json:"file_name"`
	ContentType   string `json:"content_type"`
	StepID        int64  `json:"step_id"`
	Target        string `json:"target"`
}

func (r *Registry) uploadTestCaseAttachment(ctx context.Context, args uploadTestCaseAttachmentArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if (args.FilePath == "") == (args.ContentBase64 == "") {
		return nil, fmt.Errorf("pass exactly one of file_path or content_base64")
	}
	if err := validateAttachTarget(args.Target); err != nil {
		return nil, err
	}

	var content []byte
	name := args.FileName
	if args.FilePath != "" {
		// The server reads its own disk here. That is the user's machine in
		// stdio mode, but on a shared HTTP server any client could otherwise
		// exfiltrate server files (secrets, tokens) into Allure. See localFiles.
		if !r.localFiles {
			return nil, fmt.Errorf("file_path is only allowed in local (stdio) mode; on a shared server pass content_base64 and file_name instead")
		}
		info, err := os.Stat(args.FilePath)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("file_path %q is a directory", args.FilePath)
		}
		if info.Size() > maxAttachmentBytes {
			return nil, fmt.Errorf("file is %d bytes, over the %d byte limit", info.Size(), maxAttachmentBytes)
		}
		if content, err = os.ReadFile(args.FilePath); err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}
		if name == "" {
			name = filepath.Base(args.FilePath)
		}
	} else {
		if name == "" {
			return nil, fmt.Errorf("file_name is required with content_base64")
		}
		var err error
		if content, err = base64.StdEncoding.DecodeString(strings.TrimSpace(args.ContentBase64)); err != nil {
			return nil, fmt.Errorf("content_base64 is not valid base64: %w", err)
		}
		if len(content) > maxAttachmentBytes {
			return nil, fmt.Errorf("file is %d bytes, over the %d byte limit", len(content), maxAttachmentBytes)
		}
	}

	contentType := args.ContentType
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(name))
	}
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}

	return r.uploadAndAttach(ctx, args.TestCaseID, args.StepID, args.Target, name, contentType, content)
}

type getTestCaseAttachmentContentArgs struct {
	AttachmentID int64  `json:"attachment_id"`
	SaveTo       string `json:"save_to"`
}

func (r *Registry) getTestCaseAttachmentContent(ctx context.Context, args getTestCaseAttachmentContentArgs) (any, error) {
	if args.AttachmentID <= 0 {
		return nil, fmt.Errorf("attachment_id must be positive")
	}
	if args.SaveTo != "" && !r.localFiles {
		return nil, fmt.Errorf("save_to is only allowed in local (stdio) mode")
	}
	data, contentType, err := r.allure.DownloadTestCaseAttachment(ctx, args.AttachmentID, maxAttachmentBytes)
	if err != nil {
		return nil, fmt.Errorf("download attachment: %w", err)
	}
	result := map[string]any{
		"attachment_id": args.AttachmentID,
		"content_type":  contentType,
		"size":          len(data),
	}
	if args.SaveTo != "" {
		// O_EXCL: never overwrite an existing file on the user's machine.
		f, err := os.OpenFile(args.SaveTo, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			return nil, fmt.Errorf("save file: %w", err)
		}
		_, werr := f.Write(data)
		if cerr := f.Close(); werr == nil {
			werr = cerr
		}
		if werr != nil {
			return nil, fmt.Errorf("save file: %w", werr)
		}
		result["saved_to"] = args.SaveTo
		return result, nil
	}
	if isTextContentType(contentType) && utf8.Valid(data) {
		result["content"] = string(data)
	} else {
		result["content_base64"] = base64.StdEncoding.EncodeToString(data)
	}
	return result, nil
}

type addTestCaseStepTableArgs struct {
	TestCaseID int64      `json:"test_case_id"`
	StepID     int64      `json:"step_id"`
	Rows       [][]string `json:"rows"`
	Name       string     `json:"name"`
	Target     string     `json:"target"`
}

func (r *Registry) addTestCaseStepTable(ctx context.Context, args addTestCaseStepTableArgs) (any, error) {
	if args.TestCaseID <= 0 {
		return nil, fmt.Errorf("test_case_id must be positive")
	}
	if args.StepID <= 0 {
		return nil, fmt.Errorf("step_id must be positive")
	}
	if len(args.Rows) == 0 {
		return nil, fmt.Errorf("rows must not be empty")
	}
	if err := validateAttachTarget(args.Target); err != nil {
		return nil, err
	}
	name := args.Name
	if name == "" {
		name = "Table"
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.WriteAll(args.Rows); err != nil {
		return nil, fmt.Errorf("build csv: %w", err)
	}
	// The web UI stores tables without a trailing newline (confirmed live).
	content := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))

	return r.uploadAndAttach(ctx, args.TestCaseID, args.StepID, args.Target, name, "text/csv", content)
}

func validateAttachTarget(target string) error {
	if target != "" && target != "step" && target != "expected_result" {
		return fmt.Errorf("target must be \"step\" or \"expected_result\", got %q", target)
	}
	return nil
}

// uploadAndAttach uploads the file to the test case and, when stepID is set,
// shows it in that step by creating a child scenario node pointing at the
// attachment — the same shape the web UI creates for a step's file or table
// block (confirmed live on project 408). The target step is validated before
// anything is uploaded, so a bad step id doesn't leave an orphan attachment.
func (r *Registry) uploadAndAttach(ctx context.Context, testCaseID, stepID int64, target, name, contentType string, content []byte) (any, error) {
	parentID := int64(0)
	if stepID > 0 {
		var err error
		if parentID, err = r.attachParent(ctx, testCaseID, stepID, target); err != nil {
			return nil, err
		}
	}

	att, err := r.allure.UploadTestCaseAttachment(ctx, testCaseID, name, contentType, content)
	if err != nil {
		return nil, fmt.Errorf("upload attachment: %w", err)
	}
	result := map[string]any{
		"attachment_id":  att.ID,
		"name":           att.Name,
		"content_type":   att.ContentType,
		"content_length": att.ContentLength,
	}
	if parentID == 0 {
		return result, nil
	}

	nodeID, err := r.allure.CreateTestCaseStep(ctx, allure.ScenarioStepCreateRequest{
		TestCaseID:   testCaseID,
		ParentID:     parentID,
		AttachmentID: att.ID,
	}, 0)
	if err != nil {
		result["error"] = fmt.Sprintf("uploaded to the test case, but attaching to step %d failed: %v", stepID, err)
		return result, nil
	}
	result["step_id"] = stepID
	result["node_id"] = nodeID
	return result, nil
}

// attachParent resolves the scenario node the attachment goes under: the step
// itself, or its expected-result container.
func (r *Registry) attachParent(ctx context.Context, testCaseID, stepID int64, target string) (int64, error) {
	tree, err := r.allure.GetTestCaseSteps(ctx, testCaseID)
	if err != nil {
		return 0, fmt.Errorf("look up step: %w", err)
	}
	node := stepNodeFromTree(tree, stepID)
	if node == nil {
		return 0, fmt.Errorf("step %d not found under test case %d", stepID, testCaseID)
	}
	if target != "expected_result" {
		return stepID, nil
	}
	containerID := nodeInt64(node, "expectedResultId")
	if containerID <= 0 {
		return 0, fmt.Errorf("step %d has no expected result yet — set one with update_test_case_step first, or use target \"step\"", stepID)
	}
	return containerID, nil
}
