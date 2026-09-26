package allure

import "encoding/json"

// ── Helper / sub-object DTOs ──────────────────────────────────────────────────

// CategoryDto represents a test result category.
type CategoryDto struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// TestLayerDto represents a test layer.
type TestLayerDto struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// JobRunDto represents a CI/CD job run linked to a test result.
type JobRunDto struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// IdAndNameOnlyDto is a lightweight reference with id and name.
type IdAndNameOnlyDto struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// IntegrationTypeDto represents an issue integration type.
type IntegrationTypeDto struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// RoleDto represents a user role in a project.
type RoleDto struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// StatusDto represents a named status object (for test cases).
type StatusDto struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// WorkflowRowDto represents a workflow attached to a test case.
type WorkflowRowDto struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// CustomFieldValueWithCfDto is used in create/patch to set a custom field value.
// One entry per value: {id: valueId, customField: {id: fieldId}}. An entry
// with a name and no id finds or creates the value by name (confirmed live;
// the earlier {customFieldId, values: [...]} shape 500'd).
type CustomFieldValueWithCfDto struct {
	ID          int64          `json:"id,omitempty"`
	Name        string         `json:"name,omitempty"`
	CustomField CustomFieldRef `json:"customField"`
}

// CustomFieldRef references a custom field by id.
type CustomFieldRef struct {
	ID int64 `json:"id"`
}

// ── Launch ────────────────────────────────────────────────────────────────────

type LaunchCreateRequest struct {
	Name      string            `json:"name"`
	ProjectID int64             `json:"projectId"`
	AutoClose bool              `json:"autoclose,omitempty"`
	External  bool              `json:"external,omitempty"`
	Issues    []IssueDto        `json:"issues,omitempty"`
	Links     []ExternalLinkDto `json:"links,omitempty"`
	Tags      []LaunchTag       `json:"tags,omitempty"`
}

// LaunchResponse is returned by the RS API (/api/rs/launch) and copy endpoint.
// It may include fields not present in the standard LaunchDto.
type LaunchResponse struct {
	ID               int64             `json:"id"`
	UUID             string            `json:"uuid"`
	Name             string            `json:"name"`
	Status           interface{}       `json:"status"`
	ProjectID        int64             `json:"projectId"`
	CreatedDate      int64             `json:"createdDate"`
	LastModifiedDate int64             `json:"lastModifiedDate"`
	AutoClose        bool              `json:"autoclose"`
	Closed           bool              `json:"closed"`
	External         bool              `json:"external"`
	Issues           []IssueDto        `json:"issues"`
	Links            []ExternalLinkDto `json:"links"`
	Tags             []LaunchTag       `json:"tags"`
}

type StatisticsResponse struct {
	Total   int64 `json:"total"`
	Passed  int64 `json:"passed"`
	Failed  int64 `json:"failed"`
	Broken  int64 `json:"broken"`
	Skipped int64 `json:"skipped"`
	Unknown int64 `json:"unknown"`
}

// StatisticItem is a single element from the /statistic array endpoint.
// The API returns []StatisticItem; we aggregate these into StatisticsResponse.
type StatisticItem struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

type LaunchListResponse struct {
	Content []LaunchListItem `json:"content"`
	Empty   bool             `json:"empty"`
	Last    bool             `json:"last"`
	Number  int              `json:"number"`
	Size    int              `json:"size"`
	Total   int              `json:"totalElements"`
}

type LaunchTag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type LaunchListItem struct {
	ID               int64             `json:"id"`
	UUID             string            `json:"uuid"`
	Name             string            `json:"name"`
	Status           interface{}       `json:"status"`
	ProjectID        int64             `json:"projectId"`
	StartTime        int64             `json:"startTime"`
	EndTime          int64             `json:"endTime"`
	Environment      string            `json:"environment"`
	Tags             []LaunchTag       `json:"tags"`
	CreatedDate      int64             `json:"createdDate"`
	LastModifiedDate int64             `json:"lastModifiedDate"`
	AutoClose        bool              `json:"autoclose"`
	Closed           bool              `json:"closed"`
	External         bool              `json:"external"`
	Issues           []IssueDto        `json:"issues"`
	Links            []ExternalLinkDto `json:"links"`
}

// LaunchPatchRequest is the body for PATCH /api/launch/{id} (rename, or
// toggle autoclose/external). All fields are optional — only non-nil/non-empty
// ones are sent.
type LaunchPatchRequest struct {
	Name      string `json:"name,omitempty"`
	AutoClose *bool  `json:"autoclose,omitempty"`
	External  *bool  `json:"external,omitempty"`
}

type LaunchDetailsResponse struct {
	ID               int64             `json:"id"`
	UUID             string            `json:"uuid"`
	Name             string            `json:"name"`
	Status           interface{}       `json:"status"`
	ProjectID        int64             `json:"projectId"`
	StartTime        int64             `json:"startTime"`
	EndTime          int64             `json:"endTime"`
	Environment      string            `json:"environment"`
	Tags             []LaunchTag       `json:"tags"`
	Description      string            `json:"description"`
	ReportWebUrl     string            `json:"reportWebUrl"`
	CreatedDate      int64             `json:"createdDate"`
	LastModifiedDate int64             `json:"lastModifiedDate"`
	AutoClose        bool              `json:"autoclose"`
	Closed           bool              `json:"closed"`
	External         bool              `json:"external"`
	Issues           []IssueDto        `json:"issues"`
	Links            []ExternalLinkDto `json:"links"`
}

// ── Test Results ──────────────────────────────────────────────────────────────

// TestResultParameterDto is a single parameter attached to a test result.
type TestResultParameterDto struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Hidden   bool   `json:"hidden"`
	Excluded bool   `json:"excluded"`
}

type TestResultListResponse struct {
	Content []TestResultItem `json:"content"`
	Empty   bool             `json:"empty"`
	Last    bool             `json:"last"`
	Number  int              `json:"number"`
	Size    int              `json:"size"`
	Total   int              `json:"totalElements"`
}

// TestResultItem maps the TestResultDto returned inside PageTestResultDto.
// The API uses "start"/"stop" (not "startTime"/"endTime").
type TestResultItem struct {
	ID                 int64                    `json:"id"`
	Name               string                   `json:"name"`
	Status             string                   `json:"status"`
	LaunchID           int64                    `json:"launchId"`
	TestCaseID         int64                    `json:"testCaseId"`
	ProjectID          int64                    `json:"projectId"`
	StartTime          int64                    `json:"start"`
	EndTime            int64                    `json:"stop"`
	Duration           int64                    `json:"duration"`
	Assignee           string                   `json:"assignee"`
	Muted              bool                     `json:"muted"`
	Flaky              bool                     `json:"flaky"`
	Known              bool                     `json:"known"`
	Manual             bool                     `json:"manual"`
	External           bool                     `json:"external"`
	Hidden             bool                     `json:"hidden"`
	FullName           string                   `json:"fullName"`
	Description        string                   `json:"description"`
	DescriptionHtml    string                   `json:"descriptionHtml"`
	Message            string                   `json:"message"`
	Trace              string                   `json:"trace"`
	ExpectedResult     string                   `json:"expectedResult"`
	ExpectedResultHtml string                   `json:"expectedResultHtml"`
	Precondition       string                   `json:"precondition"`
	PreconditionHtml   string                   `json:"preconditionHtml"`
	HistoryKey         string                   `json:"historyKey"`
	ScenarioKey        string                   `json:"scenarioKey"`
	ThreadId           string                   `json:"threadId"`
	HostId             string                   `json:"hostId"`
	TestedBy           string                   `json:"testedBy"`
	CreatedBy          string                   `json:"createdBy"`
	CreatedDate        int64                    `json:"createdDate"`
	LastModifiedBy     string                   `json:"lastModifiedBy"`
	LastModifiedDate   int64                    `json:"lastModifiedDate"`
	Parameters         []TestResultParameterDto `json:"parameters"`
	Tags               []TestTagDto             `json:"tags"`
	Links              []ExternalLinkDto        `json:"links"`
	Category           *CategoryDto             `json:"category"`
	Layer              *TestLayerDto            `json:"layer"`
	JobRun             *JobRunDto               `json:"jobRun"`
	RetriedBy          *IdAndNameOnlyDto        `json:"retriedBy"`
}

// TestResultDetailsResponse maps TestResultDto from GET /api/testresult/{id}.
// The API uses "start"/"stop" (not "startTime"/"endTime").
type TestResultDetailsResponse struct {
	ID                 int64                    `json:"id"`
	Name               string                   `json:"name"`
	Status             string                   `json:"status"`
	LaunchID           int64                    `json:"launchId"`
	TestCaseID         int64                    `json:"testCaseId"`
	ProjectID          int64                    `json:"projectId"`
	StartTime          int64                    `json:"start"`
	EndTime            int64                    `json:"stop"`
	Duration           int64                    `json:"duration"`
	FullName           string                   `json:"fullName"`
	Description        string                   `json:"description"`
	DescriptionHtml    string                   `json:"descriptionHtml"`
	Message            string                   `json:"message"`
	Trace              string                   `json:"trace"`
	ExpectedResult     string                   `json:"expectedResult"`
	ExpectedResultHtml string                   `json:"expectedResultHtml"`
	Precondition       string                   `json:"precondition"`
	PreconditionHtml   string                   `json:"preconditionHtml"`
	Parameters         []TestResultParameterDto `json:"parameters"`
	Assignee           string                   `json:"assignee"`
	Muted              bool                     `json:"muted"`
	Flaky              bool                     `json:"flaky"`
	Known              bool                     `json:"known"`
	Manual             bool                     `json:"manual"`
	External           bool                     `json:"external"`
	Hidden             bool                     `json:"hidden"`
	Tags               []TestTagDto             `json:"tags"`
	Links              []ExternalLinkDto        `json:"links"`
	HistoryKey         string                   `json:"historyKey"`
	ScenarioKey        string                   `json:"scenarioKey"`
	ThreadId           string                   `json:"threadId"`
	HostId             string                   `json:"hostId"`
	TestedBy           string                   `json:"testedBy"`
	CreatedBy          string                   `json:"createdBy"`
	CreatedDate        int64                    `json:"createdDate"`
	LastModifiedBy     string                   `json:"lastModifiedBy"`
	LastModifiedDate   int64                    `json:"lastModifiedDate"`
	Category           *CategoryDto             `json:"category"`
	Layer              *TestLayerDto            `json:"layer"`
	JobRun             *JobRunDto               `json:"jobRun"`
	RetriedBy          *IdAndNameOnlyDto        `json:"retriedBy"`
}

type AssignTestResultRequest struct {
	Username string `json:"username"`
}

// MuteTestResultRequest is the body for POST /api/testresult/{id}/mute.
// Name has no omitempty: the API's "mute" table has a NOT NULL constraint on
// it, and omitting the JSON key entirely (as this DTO did before) is bound
// server-side as SQL NULL, not an empty string, causing a 500 (confirmed live).
type MuteTestResultRequest struct {
	Name   string `json:"name"`
	Reason string `json:"reason,omitempty"`
}

// RunTestCaseRequest is the body for POST /api/v2/test-case/bulk/run/existing.
// There is no top-level "testCaseIds" field — the API requires a full
// TestCaseTreeSelectionDto (with its own required projectId), confirmed live:
// omitting it 409s with {"field":"selection","must not be null"} even when
// the test case is already in the launch.
type RunTestCaseRequest struct {
	LaunchId  int64                  `json:"launchId"`
	Selection TestCaseSelectionDtoV2 `json:"selection"`
}

// ── Test Cases ────────────────────────────────────────────────────────────────

type TestCaseListResponse struct {
	Content []TestCaseItem `json:"content"`
	Empty   bool           `json:"empty"`
	Last    bool           `json:"last"`
	Number  int            `json:"number"`
	Size    int            `json:"size"`
	Total   int            `json:"totalElements"`
}

type TestCaseItem struct {
	ID                 int64             `json:"id"`
	Name               string            `json:"name"`
	ProjectID          int64             `json:"projectId"`
	Status             interface{}       `json:"status"`
	AutomationStatus   interface{}       `json:"automationStatus"` // legacy; see Automated
	Automated          bool              `json:"automated"`
	Deleted            bool              `json:"deleted"`
	Editable           bool              `json:"editable"`
	External           bool              `json:"external"`
	FullName           string            `json:"fullName"`
	Hash               string            `json:"hash"`
	Description        string            `json:"description"`
	DescriptionHtml    string            `json:"descriptionHtml"`
	ExpectedResult     string            `json:"expectedResult"`
	ExpectedResultHtml string            `json:"expectedResultHtml"`
	Precondition       string            `json:"precondition"`
	PreconditionHtml   string            `json:"preconditionHtml"`
	CreatedBy          string            `json:"createdBy"`
	CreatedDate        int64             `json:"createdDate"`
	LastModifiedBy     string            `json:"lastModifiedBy"`
	LastModifiedDate   int64             `json:"lastModifiedDate"`
	Tags               []TestTagDto      `json:"tags"`
	Links              []ExternalLinkDto `json:"links"`
	TestLayer          *TestLayerDto     `json:"testLayer"`
	Workflow           *WorkflowRowDto   `json:"workflow"`
}

type TestCaseDetailsResponse struct {
	ID                 int64             `json:"id"`
	UUID               string            `json:"uuid"`
	Name               string            `json:"name"`
	ProjectID          int64             `json:"projectId"`
	Description        string            `json:"description"`
	Status             interface{}       `json:"status"`
	AutomationStatus   interface{}       `json:"automationStatus"` // legacy; see Automated
	Automated          bool              `json:"automated"`
	FullName           string            `json:"fullName"`
	Deleted            bool              `json:"deleted"`
	Editable           bool              `json:"editable"`
	External           bool              `json:"external"`
	Hash               string            `json:"hash"`
	DescriptionHtml    string            `json:"descriptionHtml"`
	ExpectedResult     string            `json:"expectedResult"`
	ExpectedResultHtml string            `json:"expectedResultHtml"`
	Precondition       string            `json:"precondition"`
	PreconditionHtml   string            `json:"preconditionHtml"`
	CreatedBy          string            `json:"createdBy"`
	CreatedDate        int64             `json:"createdDate"`
	LastModifiedBy     string            `json:"lastModifiedBy"`
	LastModifiedDate   int64             `json:"lastModifiedDate"`
	Tags               []TestTagDto      `json:"tags"`
	Links              []ExternalLinkDto `json:"links"`
	TestLayer          *TestLayerDto     `json:"testLayer"`
	Workflow           *WorkflowRowDto   `json:"workflow"`
}

// ── Projects ──────────────────────────────────────────────────────────────────

type ProjectListResponse struct {
	Content []ProjectItem `json:"content"`
	Empty   bool          `json:"empty"`
	Last    bool          `json:"last"`
	Number  int           `json:"number"`
	Size    int           `json:"size"`
	Total   int           `json:"totalElements"`
}

type ProjectItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"abbr"` // the API calls a project's code "abbr"
}

type ProjectDetailsResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"abbr"`
	Description string `json:"description"`
}

type ProjectStatsResponse struct {
	ID                 int64   `json:"id"`
	AutomatedTestCases int64   `json:"automatedTestCases"`
	ManualTestCases    int64   `json:"manualTestCases"`
	AutomationPercent  float64 `json:"automationPercent"`
	Launches           int64   `json:"launches"`
}

// ── Analytics ─────────────────────────────────────────────────────────────────

type AnalyticsResponse struct {
	Value interface{} `json:"value"`
}

type TrendData struct {
	Passed  int64 `json:"passed"`
	Failed  int64 `json:"failed"`
	Broken  int64 `json:"broken"`
	Skipped int64 `json:"skipped"`
}

// ── Test Case CRUD ────────────────────────────────────────────────────────────

// CreateTestCaseRequest maps TestCaseCreateV2Dto.
type CreateTestCaseRequest struct {
	Name           string                      `json:"name"`
	ProjectID      int64                       `json:"projectId"`
	Description    string                      `json:"description,omitempty"`
	ExpectedResult string                      `json:"expectedResult,omitempty"`
	FullName       string                      `json:"fullName,omitempty"`
	Precondition   string                      `json:"precondition,omitempty"`
	Automated      *bool                       `json:"automated,omitempty"`
	Deleted        *bool                       `json:"deleted,omitempty"`
	External       *bool                       `json:"external,omitempty"`
	StatusID       *int64                      `json:"statusId,omitempty"`
	TestLayerID    *int64                      `json:"testLayerId,omitempty"`
	WorkflowID     *int64                      `json:"workflowId,omitempty"`
	Tags           []TestTagDto                `json:"tags,omitempty"`
	Members        []MemberDto                 `json:"members,omitempty"`
	Links          []ExternalLinkDto           `json:"links,omitempty"`
	CustomFields   []CustomFieldValueWithCfDto `json:"customFields,omitempty"`
}

// UpdateTestCaseRequest maps TestCasePatchV2Dto.
//
// Text fields and lists are pointers so a caller can distinguish "don't touch"
// (nil) from "clear" (pointer to "" or an empty slice): with plain values and
// omitempty both serialize identically and the clear is silently dropped
// (confirmed live for description, precondition and tags).
type UpdateTestCaseRequest struct {
	ID             int64         `json:"id,omitempty"`
	Name           string        `json:"name,omitempty"`
	Description    *string       `json:"description,omitempty"`
	FullName       *string       `json:"fullName,omitempty"`
	Precondition   *string       `json:"precondition,omitempty"`
	ExpectedResult *string       `json:"expectedResult,omitempty"`
	Automated      *bool         `json:"automated,omitempty"`
	External       *bool         `json:"external,omitempty"`
	Deleted        *bool         `json:"deleted,omitempty"`
	StatusID       *int64        `json:"statusId,omitempty"`
	TestLayerID    *int64        `json:"testLayerId,omitempty"`
	WorkflowID     *int64        `json:"workflowId,omitempty"`
	Tags           *[]TestTagDto `json:"tags,omitempty"`
	Members        *[]MemberDto  `json:"members,omitempty"`
	// Links is a pointer so a caller can distinguish "don't touch links" (nil)
	// from "clear to no links" (pointer to an empty slice) — a plain
	// []ExternalLinkDto with omitempty would serialize both cases identically
	// (the field dropped entirely), silently no-op'ing the clear. See
	// Client.DeleteTestCaseExternalLink, which relies on this to remove the
	// last remaining link.
	Links *[]ExternalLinkDto `json:"links,omitempty"`
}

// ScenarioDto is the top-level scenario object used in PATCH /api/testcase/{id}.
type ScenarioDto struct {
	Steps []ScenarioStepDto `json:"steps"`
}

// ScenarioStepDto represents a single manual step inside a scenario.
// Type is required by the API discriminator: use "body" for a regular step,
// "expected" for an expected-result entry.
type ScenarioStepDto struct {
	Type string `json:"type"`
	Body string `json:"body,omitempty"`
}

// ── Shared DTOs ───────────────────────────────────────────────────────────────

type ExternalLinkDto struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

// ScenarioStepCreateRequest creates a scenario node. A node carries either text
// (Body) or an attachment (AttachmentID): in the web UI a step's file and
// table blocks are child nodes holding an attachment id — a table is just a
// CSV attachment (confirmed live on project 408).
type ScenarioStepCreateRequest struct {
	TestCaseID   int64  `json:"testCaseId"`
	Body         string `json:"body,omitempty"`
	ParentID     int64  `json:"parentId,omitempty"`
	AttachmentID int64  `json:"attachmentId,omitempty"`
}

type ScenarioStepPatchRequest struct {
	Body string `json:"body,omitempty"`
	// BodyJSON is the rich-text document the web UI renders. Sending plain
	// Body flattens any formatting (bold, lists, code), so when a call must
	// resend the body without changing it, send the stored BodyJSON instead
	// (confirmed live on project 408).
	BodyJSON       json.RawMessage `json:"bodyJson,omitempty"`
	ExpectedResult string          `json:"expectedResult,omitempty"`
}

type ScenarioStepCreatedResponse struct {
	CreatedStepID int64 `json:"createdStepId"`
}

type TestCaseTreeSelectionDto struct {
	ProjectID     int64     `json:"projectId"`
	TreeID        int64     `json:"treeId,omitempty"`
	Path          []int64   `json:"path,omitempty"`
	Search        string    `json:"search,omitempty"`
	LeafsInclude  []int64   `json:"leafsInclude,omitempty"`
	LeafsExclude  []int64   `json:"leafsExclude,omitempty"`
	GroupsInclude [][]int64 `json:"groupsInclude,omitempty"`
	GroupsExclude [][]int64 `json:"groupsExclude,omitempty"`
	Deleted       bool      `json:"deleted,omitempty"`
	Inverted      bool      `json:"inverted,omitempty"`
	FilterID      int64     `json:"filterId,omitempty"`
}

type TestResultTreeSelectionDto struct {
	LaunchID      int64     `json:"launchId"`
	TreeID        int64     `json:"treeId,omitempty"`
	Path          []int64   `json:"path,omitempty"`
	Search        string    `json:"search,omitempty"`
	LeafsInclude  []int64   `json:"leafsInclude,omitempty"`
	LeafsExclude  []int64   `json:"leafsExclude,omitempty"`
	GroupsInclude [][]int64 `json:"groupsInclude,omitempty"`
	GroupsExclude [][]int64 `json:"groupsExclude,omitempty"`
	Inverted      bool      `json:"inverted,omitempty"`
	FilterID      int64     `json:"filterId,omitempty"`
}

type TestTagDto struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}

type MemberDto struct {
	ID   int64    `json:"id"`
	Name string   `json:"name"`
	Role *RoleDto `json:"role,omitempty"`
}

// ── Issue ─────────────────────────────────────────────────────────────────────

// IssueDto represents an issue linked to a test case or launch.
type IssueDto struct {
	ID              int64               `json:"id,omitempty"`
	DisplayName     string              `json:"displayName,omitempty"`
	Name            string              `json:"name,omitempty"`
	URL             string              `json:"url,omitempty"`
	IntegrationID   int64               `json:"integrationId,omitempty"`
	IntegrationType *IntegrationTypeDto `json:"integrationType,omitempty"`
	Closed          bool                `json:"closed,omitempty"`
	Status          string              `json:"status,omitempty"`
	Summary         string              `json:"summary,omitempty"`
}

// ── Custom Fields ─────────────────────────────────────────────────────────────

// CustomFieldDto describes a custom field definition.
type CustomFieldDto struct {
	ID                        int64  `json:"id"`
	Name                      string `json:"name,omitempty"`
	Archived                  bool   `json:"archived,omitempty"`
	Locked                    bool   `json:"locked,omitempty"`
	Required                  bool   `json:"required,omitempty"`
	SingleSelect              bool   `json:"singleSelect,omitempty"`
	DefaultCustomFieldValueID *int64 `json:"defaultCustomFieldValueId,omitempty"`
	CreatedBy                 string `json:"createdBy,omitempty"`
	CreatedDate               int64  `json:"createdDate,omitempty"`
	LastModifiedBy            string `json:"lastModifiedBy,omitempty"`
	LastModifiedDate          int64  `json:"lastModifiedDate,omitempty"`
}

// CustomFieldValueDto describes a single value of a custom field.
type CustomFieldValueDto struct {
	ID   int64  `json:"id"`
	Name string `json:"name,omitempty"`
}

// CustomFieldWithValuesDto is used when reading or updating custom field values of a test case.
type CustomFieldWithValuesDto struct {
	CustomField CustomFieldDto        `json:"customField"`
	Values      []CustomFieldValueDto `json:"values"`
}

// IDOnlyDto references an entity by ID alone, e.g. as a nested field in a
// create request that only needs to point at an existing object.
type IDOnlyDto struct {
	ID int64 `json:"id"`
}

// CustomFieldCreateDto is the request body for creating a new custom field definition.
type CustomFieldCreateDto struct {
	Name         string `json:"name"`
	Required     bool   `json:"required"`
	SingleSelect bool   `json:"singleSelect,omitempty"`
}

// CustomFieldPatchDto is the request body for updating a custom field
// definition. Only non-nil fields are sent, since the API leaves omitted
// fields unchanged.
type CustomFieldPatchDto struct {
	Name         *string `json:"name,omitempty"`
	Required     *bool   `json:"required,omitempty"`
	SingleSelect *bool   `json:"singleSelect,omitempty"`
	Locked       *bool   `json:"locked,omitempty"`
}

// CustomFieldProjectDto describes a custom field as attached to a project,
// including its project-scoped required/locked/default settings.
type CustomFieldProjectDto struct {
	ID                        int64          `json:"id"`
	ProjectID                 int64          `json:"projectId"`
	CustomField               CustomFieldDto `json:"customField"`
	Name                      string         `json:"name,omitempty"`
	Required                  bool           `json:"required,omitempty"`
	Locked                    bool           `json:"locked,omitempty"`
	DefaultCustomFieldValueID *int64         `json:"defaultCustomFieldValueId,omitempty"`
}

// CustomFieldProjectPatchDto updates a custom field's project-scoped settings
// (required, locked, default value). Only non-nil fields are sent.
type CustomFieldProjectPatchDto struct {
	Required                  *bool  `json:"required,omitempty"`
	Locked                    *bool  `json:"locked,omitempty"`
	DefaultCustomFieldValueID *int64 `json:"defaultCustomFieldValueId,omitempty"`
}

// ListSelectionDto selects a set of IDs, optionally inverted (all except these).
type ListSelectionDto struct {
	IDs      []int64 `json:"ids"`
	Inverted bool    `json:"inverted,omitempty"`
}

// CustomFieldValueProjectCreateDto is the request body for creating a new
// custom field value option within a project.
type CustomFieldValueProjectCreateDto struct {
	CustomField IDOnlyDto `json:"customField"`
	Name        string    `json:"name"`
	Default     bool      `json:"default,omitempty"`
}

// CustomFieldValueProjectDto is a custom field value option as returned by
// the project-scoped value endpoints. Not to be confused with
// CustomFieldValueWithCfDto, which is an unrelated, differently-shaped DTO
// used for setting values on launches/test results.
type CustomFieldValueProjectDto struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name,omitempty"`
	Global      bool           `json:"global,omitempty"`
	CustomField CustomFieldDto `json:"customField"`
}

// CustomFieldValueProjectPatchDto updates a custom field value's name,
// default flag, or global flag. Only non-nil fields are sent.
type CustomFieldValueProjectPatchDto struct {
	Name    *string `json:"name,omitempty"`
	Default *bool   `json:"default,omitempty"`
	Global  *bool   `json:"global,omitempty"`
}

// ── Test Case Examples / Versions / Attachments ───────────────────────────────

// TestCaseExampleParam is a single key-value parameter in a test case example row.
type TestCaseExampleParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// TestCaseExampleDto is one row of a test case's parametrized example table, as
// returned by GET /api/testcase/{id}/example — note this differs from the POST
// request body shape (a bare [][]TestCaseExampleParam): the GET response wraps
// rows in TestCaseExamplePage and each row also carries an id and status.
type TestCaseExampleDto struct {
	ID         int64                  `json:"id"`
	Status     string                 `json:"status,omitempty"`
	Parameters []TestCaseExampleParam `json:"parameters"`
}

// TestCaseExamplePage is the paginated response of GET /api/testcase/{id}/example.
type TestCaseExamplePage struct {
	Content []TestCaseExampleDto `json:"content"`
	Last    bool                 `json:"last"`
	Number  int                  `json:"number"`
	Size    int                  `json:"size"`
	Total   int                  `json:"totalElements"`
}

// TestCaseVersionDto represents a version/snapshot of a test case.
type TestCaseVersionDto struct {
	ID          int64  `json:"id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedDate int64  `json:"createdDate,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
}

// TestCaseAttachmentDto represents an attachment on a test case.
type TestCaseAttachmentDto struct {
	ID          int64  `json:"id"`
	Name        string `json:"name,omitempty"`
	ContentType string `json:"contentType,omitempty"`
	// Size is the content length; the API has no size/createdDate fields
	// (confirmed live — they always decoded as 0).
	Size   int64 `json:"contentLength,omitempty"`
	Missed bool  `json:"missed,omitempty"`
}

// TestCaseAttachmentRowDto is one attachment returned by the upload endpoint.
type TestCaseAttachmentRowDto struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	ContentType   string `json:"contentType"`
	ContentLength int64  `json:"contentLength"`
}

// TestCaseAttachmentListResponse is the paged response for test case attachments.
type TestCaseAttachmentListResponse struct {
	Content []TestCaseAttachmentDto `json:"content"`
	Last    bool                    `json:"last"`
	Number  int                     `json:"number"`
	Size    int                     `json:"size"`
	Total   int                     `json:"totalElements"`
}

// ── Relations ─────────────────────────────────────────────────────────────────

// RelationTargetDto identifies the target of a test-case-to-test-case relation.
type RelationTargetDto struct {
	ID   int64  `json:"id"`
	Name string `json:"name,omitempty"`
}

// RelationDto represents a relation between two test cases. Type is required
// by the API (TestCaseRelationDto) — one of the TestCaseRelationTypeDto enum
// values: "related to", "clones", "is cloned by", "duplicates",
// "is duplicated by", "automates", "is automated by".
type RelationDto struct {
	ID     int64             `json:"id,omitempty"`
	Target RelationTargetDto `json:"target"`
	Type   string            `json:"type"`
}

// ── Scenario steps ────────────────────────────────────────────────────────────

// StepPositionDto is used to move or copy a scenario step.
type StepPositionDto struct {
	AfterID  int64 `json:"afterId,omitempty"`
	BeforeID int64 `json:"beforeId,omitempty"`
	ParentID int64 `json:"parentId,omitempty"`
}

// ── Test case versions ────────────────────────────────────────────────────────

// TestCaseVersionCreateRequest is the body for creating a test case version.
type TestCaseVersionCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// TestCaseVersionPatchRequest is the body for patching a test case version.
type TestCaseVersionPatchRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// ── Test keys ─────────────────────────────────────────────────────────────────

// TestKeyDto represents a test key (integration key) linked to a test case.
type TestKeyDto struct {
	ID            int64  `json:"id,omitempty"`
	IntegrationID int64  `json:"integrationId,omitempty"`
	Name          string `json:"name,omitempty"`
	URL           string `json:"url,omitempty"`
}

// ── Audit ─────────────────────────────────────────────────────────────────────

// TestCaseAuditListResponse is the paged response for test case audit log.
type TestCaseAuditListResponse struct {
	Content []map[string]any `json:"content"`
	Last    bool             `json:"last"`
	Number  int              `json:"number"`
	Size    int              `json:"size"`
	Total   int              `json:"totalElements"`
}

// ── Bulk test case DTOs ───────────────────────────────────────────────────────

type TestCaseBulkStatusDto struct {
	Selection  TestCaseSelectionDtoV2 `json:"selection"`
	StatusID   int64                  `json:"statusId"`
	WorkflowID int64                  `json:"workflowId"`
}

type TestCaseBulkTagDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	Tags      []TestTagDto           `json:"tags"`
}

type TestCaseBulkMemberDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	Members   []MemberDto            `json:"members"`
}

// TestCaseBulkEntityIdsDto is the request body for bulk remove endpoints that
// detach existing entity instances (tag/member/etc.) by their own ids rather
// than by re-sending the entity payload — e.g. POST /api/v2/test-case/bulk/tag/remove
// and POST /api/v2/test-case/bulk/member/remove. `ids` are the ids of the
// tag/member attachments to remove, not test case ids (those live in Selection).
type TestCaseBulkEntityIdsDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	IDs       []int64                `json:"ids"`
}

// ── Bulk test result DTOs ─────────────────────────────────────────────────────

type TestResultBulkAssignDto struct {
	Selection TestResultTreeSelectionDto `json:"selection"`
	Assignees []string                   `json:"assignees,omitempty"`
}

type TestResultBulkMuteDto struct {
	Selection TestResultTreeSelectionDto `json:"selection"`
	Reason    string                     `json:"reason,omitempty"`
	Name      string                     `json:"name,omitempty"`
}

type TestResultBulkResolveDto struct {
	Selection  TestResultTreeSelectionDto `json:"selection"`
	Status     string                     `json:"status"`
	CategoryID int64                      `json:"categoryId,omitempty"`
	Message    string                     `json:"message,omitempty"`
}

// TestResultBulkDto is the request body for selection-only bulk test result
// operations such as POST /api/testresult/bulk/hide.
type TestResultBulkDto struct {
	Selection TestResultTreeSelectionDto `json:"selection"`
}

// ── Launch population DTOs ────────────────────────────────────────────────────

type LaunchTestCasesAddDto struct {
	Selection       TestCaseTreeSelectionDto `json:"selection"`
	Assignees       []string                 `json:"assignees,omitempty"`
	EnvVarValueSets []interface{}            `json:"envVarValueSets,omitempty"`
	JobsMapping     []interface{}            `json:"jobsMapping,omitempty"`
	JobsParams      []interface{}            `json:"jobsParams,omitempty"`
}

type LaunchTestPlanAddDto struct {
	TestPlanID      int64         `json:"testPlanId"`
	EnvVarValueSets []interface{} `json:"envVarValueSets,omitempty"`
}

// ── Bulk test case operations ─────────────────────────────────────────────────

// TestCaseSelectionDtoV2 is the v2 test-case selection shape used by
// /api/v2/test-case/bulk/* endpoints. It is not interchangeable with
// TestCaseTreeSelectionDto (the v1 shape): field names and semantics differ
// (e.g. testCasesInclude vs leafsInclude, flat groupsInclude vs nested).
type TestCaseSelectionDtoV2 struct {
	ProjectID        int64   `json:"projectId"`
	TreeID           int64   `json:"treeId,omitempty"`
	NodeID           int64   `json:"nodeId,omitempty"`
	Search           string  `json:"search,omitempty"`
	TestCasesInclude []int64 `json:"testCasesInclude,omitempty"`
	TestCasesExclude []int64 `json:"testCasesExclude,omitempty"`
	GroupsInclude    []int64 `json:"groupsInclude,omitempty"`
	GroupsExclude    []int64 `json:"groupsExclude,omitempty"`
	Deleted          bool    `json:"deleted,omitempty"`
	Inverted         bool    `json:"inverted,omitempty"`
	FilterID         int64   `json:"filterId,omitempty"`
}

// ── Test case trees (v2) ──────────────────────────────────────────────────────
//
// A test case "tree" is a named grouping built from custom fields (e.g. the
// "Suites" tree groups by the Suite field). Its folders are custom field
// values and its leaves are test cases, so a folder only exists within a tree
// and every tree operation needs a tree id.

// TestCaseTreeDto is one named tree of a project (GET /api/v2/tree).
type TestCaseTreeDto struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ProjectID int64  `json:"projectId"`
}

// TestCaseTreePage is the paginated response of GET /api/v2/tree.
type TestCaseTreePage struct {
	Content []TestCaseTreeDto `json:"content"`
}

// TestCaseTreeNodeDto is one child of a tree node: a folder (Type "GROUP") or
// a test case (Type "LEAF"). A leaf's ID is its position in the tree, not the
// test case id, and it changes when the test case is moved.
type TestCaseTreeNodeDto struct {
	ID                 int64      `json:"id"`
	Type               string     `json:"type"`
	Name               string     `json:"name"`
	Count              int        `json:"count"`
	ParentNodeID       int64      `json:"parentNodeId"`
	CustomFieldID      int64      `json:"customFieldId"`
	CustomFieldValueID int64      `json:"customFieldValueId"`
	TestCaseID         int64      `json:"testCaseId"`
	Automated          bool       `json:"automated"`
	Status             *StatusDto `json:"status,omitempty"`
}

// TestCaseTreeNodeResponse is GET /api/v2/project/{projectId}/test-case/tree/tree-node:
// a node (the tree root when no parent is given) with one page of its children.
type TestCaseTreeNodeResponse struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ParentNodeID int64  `json:"parentNodeId"`
	Children     struct {
		Content []TestCaseTreeNodeDto `json:"content"`
		Last    bool                  `json:"last"`
		Number  int                   `json:"number"`
		Size    int                   `json:"size"`
		Total   int                   `json:"totalElements"`
	} `json:"children"`
}

// TestCaseTreeSelectionDtoV2 selects tree leaves (by leaf id) for
// POST /api/v2/test-case/tree/bulk/drag-and-drop.
type TestCaseTreeSelectionDtoV2 struct {
	ProjectID     int64   `json:"projectId"`
	TreeID        int64   `json:"treeId"`
	LeavesInclude []int64 `json:"leavesInclude"`
}

// CustomFieldValueWithCfV2Dto is a single custom-field-value assignment in the
// v2 bulk cfv shape: one row per value, with the owning custom field nested
// inside (the inverse of CustomFieldWithValuesDto's field-with-nested-values shape).
// Not to be confused with the differently-shaped CustomFieldValueWithCfDto above,
// which belongs to the test-case create/update endpoints.
type CustomFieldValueWithCfV2Dto struct {
	ID          int64          `json:"id"`
	CustomField CustomFieldDto `json:"customField"`
}

// TestCaseCfvBulkAddDto is the request body for POST /api/v2/test-case/bulk/cfv/add.
// Used instead of the v1 /api/testcase/bulk/cfv/add endpoint, which returns 500 on
// this API's backend regardless of which custom field is targeted.
type TestCaseCfvBulkAddDto struct {
	Selection TestCaseSelectionDtoV2        `json:"selection"`
	Cfv       []CustomFieldValueWithCfV2Dto `json:"cfv"`
}

// TestCaseCfvBulkRemoveDtoV2 is the request body for POST /api/v2/test-case/bulk/cfv/remove.
// Used instead of the v1 /api/testcase/bulk/cfv/remove endpoint, which reports
// success on this API's backend but silently leaves the custom field value in
// place (confirmed live: github.com/MimoJanra/TestOpsMCP/issues/18).
type TestCaseCfvBulkRemoveDtoV2 struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	IDs       []int64                `json:"ids"`
}

// BulkExternalLinkAddDto is the request body for POST /api/v2/test-case/bulk/external-link/add.
type BulkExternalLinkAddDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	Links     []ExternalLinkDto      `json:"links"`
}

// BulkIssueAddDto is the request body for POST /api/v2/test-case/bulk/issue/add.
type BulkIssueAddDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	Issues    []IssueDto             `json:"issues"`
}

// BulkIssueRemoveDto is the request body for POST /api/v2/test-case/bulk/issue/remove.
type BulkIssueRemoveDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	IDs       []int64                `json:"ids"`
}

// BulkLayerSetDto is the request body for POST /api/v2/test-case/bulk/layer/set.
type BulkLayerSetDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	LayerID   int64                  `json:"layerId"`
}

// BulkMoveDto is the request body for POST /api/v2/test-case/bulk/move.
// Maps TestCaseBulkChangeProjectDtoV2.
type BulkMoveDto struct {
	Selection   TestCaseSelectionDtoV2 `json:"selection"`
	ToProjectID int64                  `json:"toProjectId"`
	CfMapping   map[string]any         `json:"cfMapping,omitempty"`
	Strategy    string                 `json:"strategy,omitempty"`
}

// BulkDeleteDto is the request body for POST /api/v2/test-case/bulk/remove.
type BulkDeleteDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
}

// BulkRunNewLaunchDto is the request body for POST /api/v2/test-case/bulk/run/new.
type BulkRunNewLaunchDto struct {
	Selection  TestCaseSelectionDtoV2 `json:"selection"`
	LaunchName string                 `json:"launchName,omitempty"`
	Assignees  []string               `json:"assignees,omitempty"`
}

// BulkRunExistingLaunchDto is the request body for POST /api/v2/test-case/bulk/run/existing.
type BulkRunExistingLaunchDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	LaunchID  int64                  `json:"launchId"`
	Assignees []string               `json:"assignees,omitempty"`
}

// BulkCreateTestPlanDto is the request body for POST /api/v2/test-case/bulk/test-plan/create.
type BulkCreateTestPlanDto struct {
	Selection    TestCaseSelectionDtoV2 `json:"selection"`
	TestPlanName string                 `json:"testPlanName"`
}

// MuteDto is the mute reason payload nested under BulkMuteDto.Mute. The API
// requires the "mute" object itself to be present (see BulkMuteDto), and,
// mirroring MuteTestResultRequest, the DB has a NOT NULL constraint on name
// that the spec doesn't surface — always send a non-empty Name.
type MuteDto struct {
	Name   string `json:"name"`
	Reason string `json:"reason,omitempty"`
}

// BulkMuteDto is the request body for POST /api/v2/test-case/bulk/mute/add.
type BulkMuteDto struct {
	Selection TestCaseSelectionDtoV2 `json:"selection"`
	Mute      MuteDto                `json:"mute"`
}

// TestPlanDto is a test plan (GET /api/testplan/{id}).
type TestPlanDto struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	ProjectID        int64  `json:"projectId"`
	TestCasesCount   int64  `json:"testCasesCount"`
	BaseRql          string `json:"baseRql,omitempty"`
	CreatedBy        string `json:"createdBy,omitempty"`
	CreatedDate      int64  `json:"createdDate,omitempty"`
	LastModifiedDate int64  `json:"lastModifiedDate,omitempty"`
}

// TestPlanListResponse is a page of test plans.
type TestPlanListResponse struct {
	Content []TestPlanDto `json:"content"`
	Last    bool          `json:"last"`
	Number  int           `json:"number"`
	Size    int           `json:"size"`
	Total   int           `json:"totalElements"`
}

// DefectDto is a project defect (GET /api/defect/{id}).
type DefectDto struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	ProjectID        int64     `json:"projectId"`
	Closed           bool      `json:"closed"`
	CreatedDate      int64     `json:"createdDate,omitempty"`
	LastModifiedDate int64     `json:"lastModifiedDate,omitempty"`
	Issue            *IssueDto `json:"issue,omitempty"`
}

// DefectListResponse is a page of defects.
type DefectListResponse struct {
	Content []DefectDto `json:"content"`
	Last    bool        `json:"last"`
	Number  int         `json:"number"`
	Size    int         `json:"size"`
	Total   int         `json:"totalElements"`
}
