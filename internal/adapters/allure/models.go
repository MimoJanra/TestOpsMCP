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
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	Status           string `json:"status,omitempty"`
	AutomationStatus string `json:"automationStatus,omitempty"`
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
