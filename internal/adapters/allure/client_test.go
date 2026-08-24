package allure

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

// newTestServerClient builds a Client pointed at a test server. jwtPath, when
// non-empty, is where the JWT exchange is expected; the server issues a fresh
// JWT on every call unless jwtCalls is used to detect caching.
func newTestServerClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient(srv.URL, "test-api-token", 5*time.Second)
	return c, srv
}

func jwtHandler(jwtCalls *int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/uaa/oauth/token" {
			http.NotFound(w, r)
			return
		}
		if jwtCalls != nil {
			*jwtCalls++
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "fake-jwt",
			"expires_in":   3600,
		})
	}
}

func TestDoJSON_AuthHeaderAndDecode(t *testing.T) {
	var jwtCalls int
	var gotAuth string
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/uaa/oauth/token":
			jwtHandler(&jwtCalls)(w, r)
		case "/api/rs/launch":
			gotAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(LaunchResponse{ID: 42, Name: "l1"})
		default:
			http.NotFound(w, r)
		}
	})

	var result LaunchResponse
	if err := c.doJSON(context.Background(), http.MethodGet, "/api/rs/launch", nil, &result); err != nil {
		t.Fatalf("doJSON: %v", err)
	}
	if result.ID != 42 || result.Name != "l1" {
		t.Errorf("decoded result = %+v, want ID=42 Name=l1", result)
	}
	if gotAuth != "Bearer fake-jwt" {
		t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer fake-jwt")
	}
	if jwtCalls != 1 {
		t.Errorf("jwtCalls = %d, want 1", jwtCalls)
	}
}

func TestDoJSON_JWTCachedAcrossCalls(t *testing.T) {
	var jwtCalls int
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/uaa/oauth/token":
			jwtHandler(&jwtCalls)(w, r)
		default:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{})
		}
	})

	for i := 0; i < 3; i++ {
		if err := c.doJSON(context.Background(), http.MethodGet, "/api/anything", nil, nil); err != nil {
			t.Fatalf("call %d: doJSON: %v", i, err)
		}
	}
	if jwtCalls != 1 {
		t.Errorf("jwtCalls = %d, want 1 (JWT should be cached and reused)", jwtCalls)
	}
}

func TestDoJSON_NoTokenConfigured(t *testing.T) {
	c := NewClient("http://example.invalid", "", 5*time.Second)
	err := c.doJSON(context.Background(), http.MethodGet, "/api/anything", nil, nil)
	if err == nil {
		t.Fatal("expected error when no token is configured")
	}
	if !strings.Contains(err.Error(), "no token configured") {
		t.Errorf("error = %v, want mention of missing token", err)
	}
}

func TestDoJSON_NetworkError(t *testing.T) {
	c, srv := newTestServerClient(t, jwtHandler(nil))
	srv.Close() // server is now unreachable

	err := c.doJSON(context.Background(), http.MethodGet, "/api/rs/launch", nil, nil)
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestDoJSON_ErrorStatusMapsToAPIError(t *testing.T) {
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errorCode":"invalid-state","message":"bad request body"}`))
	})

	err := c.doJSON(context.Background(), http.MethodPost, "/api/rs/launch", map[string]any{"x": 1}, nil)
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
	var apiErr *APIError
	if !asAPIError(err, &apiErr) {
		t.Fatalf("error is not *APIError: %v (%T)", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Code != "invalid-state" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "invalid-state")
	}
	if !strings.Contains(apiErr.Message, "bad request body") {
		t.Errorf("Message = %q, want it to contain response body", apiErr.Message)
	}
}

func TestDoJSON_DecodeError(t *testing.T) {
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	})

	var result LaunchResponse
	err := c.doJSON(context.Background(), http.MethodGet, "/api/rs/launch", nil, &result)
	if err == nil {
		t.Fatal("expected decode error")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("error = %v, want it to mention decode failure", err)
	}
}

func TestDoRaw_CustomOkStatuses(t *testing.T) {
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})

	resp, err := c.doRaw(context.Background(), http.MethodPost, "/api/testcase/bulk/clone", nil, http.StatusOK, http.StatusAccepted, http.StatusNoContent)
	if err != nil {
		t.Fatalf("doRaw: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("StatusCode = %d, want 202", resp.StatusCode)
	}
}

func TestDoRaw_StatusNotInOkList(t *testing.T) {
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})

	_, err := c.doRaw(context.Background(), http.MethodPost, "/api/testcase/bulk/clone", nil, http.StatusOK)
	if err == nil {
		t.Fatal("expected error: 202 is not in the ok list")
	}
}

func TestURL_NormalizesMissingLeadingSlash(t *testing.T) {
	c := NewClient("http://example.invalid", "tok", time.Second)
	if got := c.url("api/foo"); got != "http://example.invalid/api/foo" {
		t.Errorf("url(%q) = %q, want %q", "api/foo", got, "http://example.invalid/api/foo")
	}
	if got := c.url("/api/foo"); got != "http://example.invalid/api/foo" {
		t.Errorf("url(%q) = %q, want %q", "/api/foo", got, "http://example.invalid/api/foo")
	}
}

func TestAPIError_ErrorMessage(t *testing.T) {
	withMsg := &APIError{StatusCode: 404, Message: "not found"}
	if got, want := withMsg.Error(), "unexpected status 404: not found"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	noMsg := &APIError{StatusCode: 500}
	if got, want := noMsg.Error(), "unexpected status 500"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestErrFromResponse_ParsesErrorCodeVariants(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"code", `{"code":"a"}`, "a"},
		{"errorCode", `{"errorCode":"b"}`, "b"},
		{"errorType", `{"errorType":"c"}`, "c"},
		{"non-json body", "plain text error", ""},
		{"empty body", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			rec.WriteHeader(http.StatusInternalServerError)
			_, _ = rec.WriteString(tc.body)
			resp := rec.Result()
			err := errFromResponse(resp)
			var apiErr *APIError
			if !asAPIError(err, &apiErr) {
				t.Fatalf("not *APIError: %v", err)
			}
			if apiErr.Code != tc.want {
				t.Errorf("Code = %q, want %q", apiErr.Code, tc.want)
			}
		})
	}
}

// asAPIError mimics errors.As for *APIError without importing errors just for one check.
func asAPIError(err error, target **APIError) bool {
	if ae, ok := err.(*APIError); ok {
		*target = ae
		return true
	}
	return false
}

// --- Representative public-method tests: verify URL/query/path construction ---

func TestListLaunches_BuildsQueryParams(t *testing.T) {
	var gotQuery url.Values
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(LaunchListResponse{Total: 1})
	})

	result, err := c.ListLaunches(context.Background(), 7, 2, 25)
	if err != nil {
		t.Fatalf("ListLaunches: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total = %d, want 1", result.Total)
	}
	if got := gotQuery.Get("projectId"); got != "7" {
		t.Errorf("projectId = %q, want 7", got)
	}
	if got := gotQuery.Get("page"); got != "2" {
		t.Errorf("page = %q, want 2", got)
	}
	if got := gotQuery.Get("size"); got != "25" {
		t.Errorf("size = %q, want 25", got)
	}
}

func TestDeleteTestCase_PathParamAndNoContent(t *testing.T) {
	var gotMethod, gotPath string
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		gotMethod, gotPath = r.Method, r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.DeleteTestCase(context.Background(), 99); err != nil {
		t.Fatalf("DeleteTestCase: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
	if gotPath != "/api/testcase/99" {
		t.Errorf("path = %q, want /api/testcase/99", gotPath)
	}
}

func TestCreateLaunch_SendsJSONBody(t *testing.T) {
	var gotBody map[string]any
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(LaunchResponse{ID: 5, Name: "created"})
	})

	result, err := c.CreateLaunch(context.Background(), 3, "my launch")
	if err != nil {
		t.Fatalf("CreateLaunch: %v", err)
	}
	if result.ID != 5 {
		t.Errorf("ID = %d, want 5", result.ID)
	}
	if gotBody["name"] != "my launch" || gotBody["projectId"] != float64(3) {
		t.Errorf("request body = %+v, want name=my launch projectId=3", gotBody)
	}
}

func TestGetLaunchStatistics_AggregatesItems(t *testing.T) {
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]StatisticItem{
			{Status: "passed", Count: 3},
			{Status: "failed", Count: 1},
			{Status: "weird", Count: 2},
		})
	})

	stats, err := c.GetLaunchStatistics(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetLaunchStatistics: %v", err)
	}
	if stats.Passed != 3 || stats.Failed != 1 || stats.Unknown != 2 || stats.Total != 6 {
		t.Errorf("stats = %+v, want Passed=3 Failed=1 Unknown=2 Total=6", stats)
	}
}

func TestCloseLaunch_NoContentIsSuccess(t *testing.T) {
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	if err := c.CloseLaunch(context.Background(), 1); err != nil {
		t.Fatalf("CloseLaunch: %v", err)
	}
}

func TestListProjects_Success(t *testing.T) {
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		if r.URL.Path != "/api/project" {
			t.Errorf("path = %q, want /api/project", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ProjectListResponse{})
	})

	if _, err := c.ListProjects(context.Background(), 0, 10); err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
}

// zeroArg builds a harmless placeholder value for a method parameter type, used
// by the generic sweep below to invoke every exported Client method without
// hand-writing one test per method.
func zeroArg(t reflect.Type) reflect.Value {
	if t == reflect.TypeOf((*http.Request)(nil)) {
		return reflect.ValueOf(httptest.NewRequest(http.MethodGet, "http://example.invalid/", nil))
	}
	switch t.Kind() {
	case reflect.Slice:
		return reflect.MakeSlice(t, 0, 0)
	case reflect.Map:
		return reflect.MakeMap(t)
	case reflect.Ptr:
		return reflect.New(t.Elem())
	case reflect.String:
		return reflect.ValueOf("x").Convert(t)
	case reflect.Bool:
		return reflect.ValueOf(false)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(int64(1)).Convert(t)
	case reflect.Struct:
		return reflect.New(t).Elem()
	default:
		return reflect.Zero(t)
	}
}

// hasSliceReturn reports whether any of the method's return values is a slice
// or array, which is how we decide whether the stub server should answer with
// a JSON array ("[]") or a JSON object ("{}") so json.Decode succeeds.
func hasSliceReturn(mt reflect.Type) bool {
	for i := 0; i < mt.NumOut(); i++ {
		if k := mt.Out(i).Kind(); k == reflect.Slice || k == reflect.Array {
			return true
		}
	}
	return false
}

// TestAllClientMethods_ExecuteAgainstStubServer drives every exported Client
// method at least once against a stub server that always answers 200 OK. It
// does not assert on results — its job is to exercise the thin per-endpoint
// wrappers (URL/body construction + the doJSON/doRaw call) that dedicated
// tests above don't each cover individually, now that the client is ~110
// near-identical wrapper methods around a couple of shared helpers.
func TestAllClientMethods_ExecuteAgainstStubServer(t *testing.T) {
	var wantArray bool
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		if wantArray {
			_, _ = w.Write([]byte("[]"))
		} else {
			_, _ = w.Write([]byte("{}"))
		}
	})

	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	v := reflect.ValueOf(c)
	ct := v.Type()
	for i := 0; i < ct.NumMethod(); i++ {
		m := ct.Method(i)
		if m.PkgPath != "" {
			continue // unexported method
		}
		mv := v.Method(i)
		mt := mv.Type()
		wantArray = hasSliceReturn(mt)

		args := make([]reflect.Value, mt.NumIn())
		for j := 0; j < mt.NumIn(); j++ {
			pt := mt.In(j)
			if j == 0 && pt == ctxType {
				args[j] = reflect.ValueOf(context.Background())
				continue
			}
			args[j] = zeroArg(pt)
		}

		name := m.Name
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("method %s panicked: %v", name, r)
				}
			}()
			mv.Call(args)
		}()
	}
}

func TestMergeLaunches_MultiLineMarshaledBody(t *testing.T) {
	var gotBody map[string]any
	c, _ := newTestServerClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/uaa/oauth/token" {
			jwtHandler(nil)(w, r)
			return
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 77})
	})

	id, err := c.MergeLaunches(context.Background(), []int64{1, 2, 3}, "merged")
	if err != nil {
		t.Fatalf("MergeLaunches: %v", err)
	}
	if id != 77 {
		t.Errorf("id = %d, want 77", id)
	}
	if gotBody["name"] != "merged" {
		t.Errorf("request body name = %v, want merged", gotBody["name"])
	}
}
