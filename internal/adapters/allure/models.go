package allure

type LaunchCreateRequest struct {
	Name      string `json:"name"`
	ProjectID int64  `json:"projectId"`
}

type LaunchResponse struct {
	ID     int64       `json:"id"`
	UUID   string      `json:"uuid"`
	Name   string      `json:"name"`
	Status interface{} `json:"status"`
}

type StatisticsResponse struct {
	Total  int64 `json:"total"`
	Passed int64 `json:"passed"`
	Failed int64 `json:"failed"`
	Broken int64 `json:"broken"`
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
	ID          int64       `json:"id"`
	UUID        string      `json:"uuid"`
	Name        string      `json:"name"`
	Status      interface{} `json:"status"`
	ProjectID   int64       `json:"projectId"`
	StartTime   int64       `json:"startTime"`
	EndTime     int64       `json:"endTime"`
	Environment string      `json:"environment"`
	Tags        []LaunchTag `json:"tags"`
}

type LaunchDetailsResponse struct {
	ID           int64       `json:"id"`
	UUID         string      `json:"uuid"`
	Name         string      `json:"name"`
	Status       interface{} `json:"status"`
	ProjectID    int64       `json:"projectId"`
	StartTime    int64       `json:"startTime"`
	EndTime      int64       `json:"endTime"`
	Environment  string      `json:"environment"`
	Tags         []LaunchTag `json:"tags"`
	Description  string      `json:"description"`
	ReportWebUrl string      `json:"reportWebUrl"`
}

type TestResultListResponse struct {
	Content []TestResultItem `json:"content"`
	Empty   bool             `json:"empty"`
	Last    bool             `json:"last"`
	Number  int              `json:"number"`
	Size    int              `json:"size"`
	Total   int              `json:"totalElements"`
}

type TestResultItem struct {
	ID        int64       `json:"id"`
	Name      string      `json:"name"`
	Status    interface{} `json:"status"`
	LaunchID  int64       `json:"launchId"`
	StartTime int64       `json:"startTime"`
	EndTime   int64       `json:"endTime"`
	Duration  int64       `json:"duration"`
}

type TestResultDetailsResponse struct {
	ID          int64       `json:"id"`
	UUID        string      `json:"uuid"`
	Name        string      `json:"name"`
	Status      interface{} `json:"status"`
	LaunchID    int64       `json:"launchId"`
	StartTime   int64       `json:"startTime"`
	EndTime     int64       `json:"endTime"`
	Duration    int64       `json:"duration"`
	FullName    string      `json:"fullName"`
	Description string      `json:"description"`
	Parameters  string      `json:"parameters"`
}

type AssignTestResultRequest struct {
	Username string `json:"username"`
}

type MuteTestResultRequest struct {
	Reason string `json:"reason"`
}

type RunTestCaseRequest struct {
	TestCaseIds []int64 `json:"testCaseIds"`
	LaunchId    int64   `json:"launchId"`
}

type TestCaseListResponse struct {
	Content []TestCaseItem `json:"content"`
	Empty   bool           `json:"empty"`
	Last    bool           `json:"last"`
	Number  int            `json:"number"`
	Size    int            `json:"size"`
	Total   int            `json:"totalElements"`
}

type TestCaseItem struct {
	ID               int64       `json:"id"`
	Name             string      `json:"name"`
	ProjectID        int64       `json:"projectId"`
	Status           interface{} `json:"status"`
	AutomationStatus interface{} `json:"automationStatus"`
}

type TestCaseDetailsResponse struct {
	ID               int64       `json:"id"`
	UUID             string      `json:"uuid"`
	Name             string      `json:"name"`
	ProjectID        int64       `json:"projectId"`
	Description      string      `json:"description"`
	Status           interface{} `json:"status"`
	AutomationStatus interface{} `json:"automationStatus"`
	FullName         string      `json:"fullName"`
}

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
	Code string `json:"code"`
}

type ProjectDetailsResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type ProjectStatsResponse struct {
	ID                 int64   `json:"id"`
	AutomatedTestCases int64   `json:"automatedTestCases"`
	ManualTestCases    int64   `json:"manualTestCases"`
	AutomationPercent  float64 `json:"automationPercent"`
	Launches           int64   `json:"launches"`
}

type AnalyticsResponse struct {
	Value interface{} `json:"value"`
}

type TrendData struct {
	Passed  int64 `json:"passed"`
	Failed  int64 `json:"failed"`
	Broken  int64 `json:"broken"`
	Skipped int64 `json:"skipped"`
}

type CreateTestCaseRequest struct {
	Name             string `json:"name"`
	ProjectID        int64  `json:"projectId"`
	Description      string `json:"description,omitempty"`
	Status           string `json:"status,omitempty"`
	AutomationStatus string `json:"automationStatus,omitempty"`
}

type UpdateTestCaseRequest struct {
	Name            string            `json:"name,omitempty"`
	Description     string            `json:"description,omitempty"`
	FullName        string            `json:"fullName,omitempty"`
	Precondition    string            `json:"precondition,omitempty"`
	ExpectedResult  string            `json:"expectedResult,omitempty"`
	Automated       *bool             `json:"automated,omitempty"`
	External        *bool             `json:"external,omitempty"`
	Deleted         *bool             `json:"deleted,omitempty"`
	StatusID        *int64            `json:"statusId,omitempty"`
	TestLayerID     *int64            `json:"testLayerId,omitempty"`
	WorkflowID      *int64            `json:"workflowId,omitempty"`
	Tags            []TestTagDto      `json:"tags,omitempty"`
	Members         []MemberDto       `json:"members,omitempty"`
	Links           []ExternalLinkDto `json:"links,omitempty"`
	ManualScenario  map[string]any    `json:"manualScenario,omitempty"`
}

type ExternalLinkDto struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
	URL  string `json:"url,omitempty"`
}

type ScenarioStepCreateRequest struct {
	TestCaseID int64  `json:"testCaseId"`
	Body       string `json:"body,omitempty"`
	ParentID   int64  `json:"parentId,omitempty"`
}

type ScenarioStepPatchRequest struct {
	Body           string `json:"body,omitempty"`
	ExpectedResult string `json:"expectedResult,omitempty"`
}

type ScenarioStepCreatedResponse struct {
	CreatedStepID int64 `json:"createdStepId"`
}

type TestCaseTreeSelectionDto struct {
	ProjectID      int64     `json:"projectId"`
	TreeID         int64     `json:"treeId,omitempty"`
	Path           []int64   `json:"path,omitempty"`
	Search         string    `json:"search,omitempty"`
	LeafsInclude   []int64   `json:"leafsInclude,omitempty"`
	LeafsExclude   []int64   `json:"leafsExclude,omitempty"`
	GroupsInclude  [][]int64 `json:"groupsInclude,omitempty"`
	GroupsExclude  [][]int64 `json:"groupsExclude,omitempty"`
	Deleted        bool      `json:"deleted,omitempty"`
	Inverted       bool      `json:"inverted,omitempty"`
	FilterID       int64     `json:"filterId,omitempty"`
}

type TestResultTreeSelectionDto struct {
	LaunchID       int64     `json:"launchId"`
	TreeID         int64     `json:"treeId,omitempty"`
	Path           []int64   `json:"path,omitempty"`
	Search         string    `json:"search,omitempty"`
	LeafsInclude   []int64   `json:"leafsInclude,omitempty"`
	LeafsExclude   []int64   `json:"leafsExclude,omitempty"`
	GroupsInclude  [][]int64 `json:"groupsInclude,omitempty"`
	GroupsExclude  [][]int64 `json:"groupsExclude,omitempty"`
	Inverted       bool      `json:"inverted,omitempty"`
	FilterID       int64     `json:"filterId,omitempty"`
}

type TestTagDto struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type MemberDto struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TestCaseBulkStatusDto struct {
	Selection  TestCaseTreeSelectionDto `json:"selection"`
	StatusID   int64                    `json:"statusId"`
	WorkflowID int64                    `json:"workflowId"`
}

type TestCaseBulkTagDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
	Tags      []TestTagDto             `json:"tags"`
}

type TestCaseBulkMemberDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
	Members   []MemberDto              `json:"members"`
}

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
	Selection TestResultTreeSelectionDto `json:"selection"`
	Issues    []interface{}              `json:"issues,omitempty"`
}

type LaunchTestCasesAddDto struct {
	Selection        TestCaseTreeSelectionDto `json:"selection"`
	Assignees        []string                 `json:"assignees,omitempty"`
	EnvVarValueSets  []interface{}            `json:"envVarValueSets,omitempty"`
	JobsMapping      []interface{}            `json:"jobsMapping,omitempty"`
	JobsParams       []interface{}            `json:"jobsParams,omitempty"`
}

type LaunchTestPlanAddDto struct {
	TestPlanID      int64         `json:"testPlanId"`
	EnvVarValueSets []interface{} `json:"envVarValueSets,omitempty"`
}

// CustomFieldDto describes a custom field definition.
type CustomFieldDto struct {
	ID   int64  `json:"id"`
	Name string `json:"name,omitempty"`
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

// IssueDto represents an issue linked to a test case.
type IssueDto struct {
	ID            int64  `json:"id,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	URL           string `json:"url,omitempty"`
	IntegrationID int64  `json:"integrationId,omitempty"`
	Closed        bool   `json:"closed,omitempty"`
}

// TestCaseExampleParam is a single key-value parameter in a test case example row.
type TestCaseExampleParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
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
	Size        int64  `json:"size,omitempty"`
	CreatedDate int64  `json:"createdDate,omitempty"`
}

// TestCaseAttachmentListResponse is the paged response for test case attachments.
type TestCaseAttachmentListResponse struct {
	Content []TestCaseAttachmentDto `json:"content"`
	Last    bool                    `json:"last"`
	Number  int                     `json:"number"`
	Size    int                     `json:"size"`
	Total   int                     `json:"totalElements"`
}

// RelationTargetDto identifies the target of a test-case-to-test-case relation.
type RelationTargetDto struct {
	ID   int64  `json:"id"`
	Name string `json:"name,omitempty"`
}

// RelationDto represents a relation between two test cases.
type RelationDto struct {
	ID     int64             `json:"id,omitempty"`
	Target RelationTargetDto `json:"target"`
}

// StepPositionDto is used to move or copy a scenario step.
type StepPositionDto struct {
	AfterID  int64 `json:"afterId,omitempty"`
	BeforeID int64 `json:"beforeId,omitempty"`
	ParentID int64 `json:"parentId,omitempty"`
}

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

// TestKeyDto represents a test key (integration key) linked to a test case.
type TestKeyDto struct {
	ID            int64  `json:"id,omitempty"`
	IntegrationID int64  `json:"integrationId,omitempty"`
	Name          string `json:"name,omitempty"`
	URL           string `json:"url,omitempty"`
}

// TestCaseAuditListResponse is the paged response for test case audit log.
type TestCaseAuditListResponse struct {
	Content []map[string]any `json:"content"`
	Last    bool             `json:"last"`
	Number  int              `json:"number"`
	Size    int              `json:"size"`
	Total   int              `json:"totalElements"`
}

// ── Bulk DTOs ─────────────────────────────────────────────────────────────────

// BulkCfvAddDto is the request body for POST /api/testcase/bulk/cfv/add.
type BulkCfvAddDto struct {
	Selection TestCaseTreeSelectionDto   `json:"selection"`
	Cfv       []CustomFieldWithValuesDto `json:"cfv"`
}

// BulkCfvRemoveDto is the request body for POST /api/testcase/bulk/cfv/remove.
type BulkCfvRemoveDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
	IDs       []int64                  `json:"ids"`
}

// BulkExternalLinkAddDto is the request body for POST /api/testcase/bulk/externallink/add.
type BulkExternalLinkAddDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
	Links     []ExternalLinkDto        `json:"links"`
}

// BulkIssueAddDto is the request body for POST /api/testcase/bulk/issue/add.
type BulkIssueAddDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
	Issues    []IssueDto               `json:"issues"`
}

// BulkIssueRemoveDto is the request body for POST /api/testcase/bulk/issue/remove.
type BulkIssueRemoveDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
	IDs       []int64                  `json:"ids"`
}

// BulkLayerSetDto is the request body for POST /api/testcase/bulk/layer/set.
type BulkLayerSetDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
	LayerID   int64                    `json:"layerId"`
}

// BulkMoveDto is the request body for POST /api/testcase/bulk/move.
type BulkMoveDto struct {
	Selection   TestCaseTreeSelectionDto `json:"selection"`
	ToProjectID int64                    `json:"toProjectId"`
}

// BulkDeleteDto is the request body for POST /api/testcase/bulk/remove.
type BulkDeleteDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
}

// BulkRunNewLaunchDto is the request body for POST /api/testcase/bulk/run/new.
type BulkRunNewLaunchDto struct {
	Selection  TestCaseTreeSelectionDto `json:"selection"`
	LaunchName string                   `json:"launchName,omitempty"`
	Assignees  []string                 `json:"assignees,omitempty"`
}

// BulkRunExistingLaunchDto is the request body for POST /api/testcase/bulk/run/existing.
type BulkRunExistingLaunchDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
	LaunchID  int64                    `json:"launchId"`
	Assignees []string                 `json:"assignees,omitempty"`
}

// BulkCreateTestPlanDto is the request body for POST /api/testcase/bulk/testplan/create.
type BulkCreateTestPlanDto struct {
	Selection    TestCaseTreeSelectionDto `json:"selection"`
	TestPlanName string                   `json:"testPlanName"`
}

// BulkMuteDto is the request body for POST /api/testcase/bulk/mute/add.
type BulkMuteDto struct {
	Selection TestCaseTreeSelectionDto `json:"selection"`
	Mute      map[string]any           `json:"mute,omitempty"`
}
