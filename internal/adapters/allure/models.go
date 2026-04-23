package allure

type LaunchCreateRequest struct {
	Name      string `json:"name"`
	ProjectID int64  `json:"projectId"`
}

type LaunchResponse struct {
	ID     int64  `json:"id"`
	UUID   string `json:"uuid"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type LaunchStatusResponse struct {
	Status string `json:"status"`
}

type StatisticsResponse struct {
	Total  int64 `json:"total"`
	Passed int64 `json:"passed"`
	Failed int64 `json:"failed"`
	Broken int64 `json:"broken"`
}
