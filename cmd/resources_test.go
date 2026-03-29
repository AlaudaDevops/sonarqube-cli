package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sqclient "github.com/tektoncd/operator/tools/sonarqube-cli/pkg/client"
	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
)

func TestRunCreate_RequiresTokenFileBeforeAPICalls(t *testing.T) {
	oldConfigFile := configFile
	oldPlugin := plugin
	oldTokenFile := tokenFile
	oldStateFile := stateFile
	t.Cleanup(func() {
		configFile = oldConfigFile
		plugin = oldPlugin
		tokenFile = oldTokenFile
		stateFile = oldStateFile
	})

	configFile = "does-not-matter.yaml"
	plugin = "tektoncd"
	tokenFile = ""
	stateFile = "does-not-matter.state"

	err := runCreate(createCmd, nil)
	if err == nil {
		t.Fatal("runCreate() error = nil, want token-file validation error")
	}
	if !strings.Contains(err.Error(), "--token-file is required") {
		t.Fatalf("runCreate() error = %v, want token-file validation error", err)
	}
}

func TestRunCreate_RejectsExistingStateFileBeforeLoadingConfig(t *testing.T) {
	dir := t.TempDir()
	existingStateFile := filepath.Join(dir, "resource.state")
	if err := os.WriteFile(existingStateFile, []byte("existing-state"), 0600); err != nil {
		t.Fatalf("WriteFile(state) error = %v", err)
	}

	oldConfigFile := configFile
	oldPlugin := plugin
	oldTokenFile := tokenFile
	oldStateFile := stateFile
	oldResolvedConfig := resolvedConfig
	oldOutputTemplate := outputTemplate
	oldOutputFile := outputFile
	t.Cleanup(func() {
		configFile = oldConfigFile
		plugin = oldPlugin
		tokenFile = oldTokenFile
		stateFile = oldStateFile
		resolvedConfig = oldResolvedConfig
		outputTemplate = oldOutputTemplate
		outputFile = oldOutputFile
	})

	configFile = filepath.Join(dir, "does-not-matter.yaml")
	plugin = "tektoncd"
	tokenFile = filepath.Join(dir, "token.env")
	stateFile = existingStateFile
	resolvedConfig = ""
	outputTemplate = ""
	outputFile = ""

	err := runCreate(createCmd, nil)
	if err == nil {
		t.Fatal("runCreate() error = nil, want existing state-file validation error")
	}
	if !strings.Contains(err.Error(), "--state-file already exists") {
		t.Fatalf("runCreate() error = %v, want existing state-file validation error", err)
	}
}

func TestCreateCommand_RequiresTokenFileFlag(t *testing.T) {
	flag := createCmd.Flag("token-file")
	if flag == nil {
		t.Fatal("createCmd token-file flag = nil")
	}
	if _, ok := createCmd.Flag("token-file").Annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
		t.Fatal("createCmd token-file flag is not marked required")
	}
	if _, ok := createCmd.Flag("state-file").Annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
		t.Fatal("createCmd state-file flag is not marked required")
	}
	if _, ok := cleanupCmd.Flag("state-file").Annotations["cobra_annotation_bash_completion_one_required_flag"]; !ok {
		t.Fatal("cleanupCmd state-file flag is not marked required")
	}
}

func TestFilterMatchesTemplatedPluginName(t *testing.T) {
	res := config.TempResource{
		PluginName: "${PLUGIN_NAME}",
	}

	resolvedPluginName, err := config.ReplaceVariables(res.PluginName, "task-1", "tektoncd")
	if err != nil {
		t.Fatalf("ReplaceVariables() error = %v", err)
	}
	if resolvedPluginName != "tektoncd" {
		t.Fatalf("ReplaceVariables() = %q, want %q", resolvedPluginName, "tektoncd")
	}
}

func TestValidateCreateOutputTargets_RejectsExistingPaths(t *testing.T) {
	dir := t.TempDir()

	testCases := []struct {
		name string
		flag string
		path string
	}{
		{name: "token file", flag: "--token-file", path: filepath.Join(dir, "token.env")},
		{name: "resolved config", flag: "--resolved-config", path: filepath.Join(dir, "resolved.yaml")},
		{name: "output file", flag: "--output-file", path: filepath.Join(dir, "output.txt")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := os.WriteFile(tc.path, []byte("existing"), 0600); err != nil {
				t.Fatalf("WriteFile(%s) error = %v", tc.name, err)
			}

			err := validateCreateOutputTargets(outputTarget{flag: tc.flag, path: tc.path})
			if err == nil {
				t.Fatalf("validateCreateOutputTargets() error = nil, want existing %s validation error", tc.flag)
			}
			if !strings.Contains(err.Error(), tc.flag+" already exists") {
				t.Fatalf("validateCreateOutputTargets() error = %v, want %s already exists", err, tc.flag)
			}
		})
	}
}

func TestCreateResources_RollbackDeletesOnlyCreatedResourcesOnConflict(t *testing.T) {
	var deleteGroupCalls int
	var deleteUserCalls int
	var deleteProjectCalls int
	var deleteTemplateCalls int
	var revokeTokenCalls int

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/authorizations/groups", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("/api/users/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"errors":[{"msg":"User already exists"}]}`)
	})
	mux.HandleFunc("/api/v2/authorizations/groups/test-group", func(w http.ResponseWriter, r *http.Request) {
		deleteGroupCalls++
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/users/deactivate", func(w http.ResponseWriter, r *http.Request) {
		deleteUserCalls++
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/projects/delete", func(w http.ResponseWriter, r *http.Request) {
		deleteProjectCalls++
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/permissions/delete_template", func(w http.ResponseWriter, r *http.Request) {
		deleteTemplateCalls++
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/user_tokens/revoke", func(w http.ResponseWriter, r *http.Request) {
		revokeTokenCalls++
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := sqclient.New(server.URL, "test-token")
	res := config.TempResource{
		PluginName: "tektoncd",
		Group: config.Group{
			Name:        "test-group",
			Description: "Temporary group",
		},
		User: config.User{
			Login:    "test-user",
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "password",
		},
		Projects: []config.Project{
			{Key: "test-project", Name: "Test Project"},
		},
		PermissionTemplate: config.PermissionTemplate{
			Name: "test-template",
		},
	}

	token, plan, err := createResources(c, res, "task-1")
	if token != "" {
		t.Fatalf("createResources() token = %q, want empty", token)
	}
	if plan == nil {
		t.Fatal("createResources() plan = nil")
	}
	if !errors.Is(err, sqclient.ErrAlreadyExists) {
		t.Fatalf("createResources() error = %v, want ErrAlreadyExists", err)
	}
	if deleteGroupCalls != 1 {
		t.Fatalf("deleteGroupCalls = %d, want 1", deleteGroupCalls)
	}
	if deleteUserCalls != 0 || deleteProjectCalls != 0 || deleteTemplateCalls != 0 || revokeTokenCalls != 0 {
		t.Fatalf("unexpected cleanup calls: user=%d project=%d template=%d token=%d", deleteUserCalls, deleteProjectCalls, deleteTemplateCalls, revokeTokenCalls)
	}
}

func TestCleanupOnError_ReportsCleanupFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/authorizations/groups/test-group", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"errors":[{"msg":"delete failed"}]}`)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := sqclient.New(server.URL, "test-token")
	res := config.TempResource{
		PluginName: "tektoncd",
		Group: config.Group{
			Name: "test-group",
		},
	}
	plan := &cleanupPlan{deleteGroup: true}

	err := cleanupOnError(c, res, plan, "task-1", errors.New("write failed"))
	if err == nil {
		t.Fatal("cleanupOnError() error = nil, want cleanup failure")
	}
	if !strings.Contains(err.Error(), "write failed") || !strings.Contains(err.Error(), "cleanup failed") {
		t.Fatalf("cleanupOnError() error = %v, want original and cleanup failure details", err)
	}
}

func TestSelectTempResource_RejectsMultipleMatches(t *testing.T) {
	cfg := &config.Config{
		SonarQube: config.SonarQubeConfig{
			TempResources: []config.TempResource{
				{PluginName: "tektoncd", Group: config.Group{Name: "group-1"}, User: config.User{Login: "user-1"}},
				{PluginName: "${PLUGIN_NAME}", Group: config.Group{Name: "group-2"}, User: config.User{Login: "user-2"}},
			},
		},
	}

	_, _, err := selectTempResource(cfg, "tektoncd", "task-1")
	if err == nil || !strings.Contains(err.Error(), "multiple temp_resources matched plugin") {
		t.Fatalf("selectTempResource() error = %v, want duplicate match error", err)
	}
}

func TestResourceState_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state", "resource-state.yaml")
	want := resourceState{
		Version:            1,
		Plugin:             "tektoncd",
		Endpoint:           "https://sonarqube.example.com",
		TaskRunID:          "task-1",
		GroupName:          "test-group",
		UserLogin:          "test-user",
		ProjectKeys:        []string{"project-a"},
		PermissionTemplate: "test-template",
		TokenName:          "test-token-task-1",
	}

	if err := writeResourceState(path, want); err != nil {
		t.Fatalf("writeResourceState() error = %v", err)
	}

	got, err := loadResourceState(path)
	if err != nil {
		t.Fatalf("loadResourceState() error = %v", err)
	}
	if got.Plugin != want.Plugin || got.Endpoint != want.Endpoint || got.TaskRunID != want.TaskRunID || got.GroupName != want.GroupName || got.TokenName != want.TokenName {
		t.Fatalf("state round trip mismatch: got %+v want %+v", got, want)
	}
}

func TestRunCleanup_UsesStateFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	statePath := filepath.Join(dir, "state.yaml")
	oldConfigFile := configFile
	oldPlugin := plugin
	oldStateFile := stateFile
	t.Cleanup(func() {
		configFile = oldConfigFile
		plugin = oldPlugin
		stateFile = oldStateFile
	})

	cfgData := []byte(`sonarqube:
  endpoint: https://sonarqube.example.com
  manager:
    username: manager
    token: token
  temp_resources:
    - plugin_name: unrelated
      group:
        name: ""
      user:
        login: ""
`)
	if err := os.WriteFile(cfgPath, cfgData, 0600); err != nil {
		t.Fatalf("WriteFile(config) error = %v", err)
	}
	configFile = cfgPath
	stateFile = statePath
	plugin = "tektoncd"

	mux := http.NewServeMux()
	mux.HandleFunc("/api/user_tokens/revoke", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	mux.HandleFunc("/api/projects/delete", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.PostForm.Get("project"); got != "state-project" {
			t.Fatalf("project delete target = %q, want state-project", got)
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/permissions/delete_template", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.PostForm.Get("templateName"); got != "state-template" {
			t.Fatalf("template delete target = %q, want state-template", got)
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/users/deactivate", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm() error = %v", err)
		}
		if got := r.PostForm.Get("login"); got != "state-user" {
			t.Fatalf("user delete target = %q, want state-user", got)
		}
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/v2/authorizations/groups/state-group", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	if err := writeResourceState(statePath, resourceState{
		Version:            1,
		Plugin:             "tektoncd",
		Endpoint:           strings.TrimSuffix(server.URL, "/"),
		TaskRunID:          "task-1",
		GroupName:          "state-group",
		UserLogin:          "state-user",
		ProjectKeys:        []string{"state-project"},
		PermissionTemplate: "state-template",
		TokenName:          "test-token-task-1",
	}); err != nil {
		t.Fatalf("writeResourceState(state) error = %v", err)
	}

	t.Setenv("SONARQUBE_ALLOW_HTTP", "true")
	cfgData = []byte(fmt.Sprintf(`sonarqube:
  endpoint: %s
  manager:
    username: manager
    token: token
  temp_resources:
    - plugin_name: unrelated
      group:
        name: ""
      user:
        login: ""
`, server.URL))
	if err := os.WriteFile(cfgPath, cfgData, 0600); err != nil {
		t.Fatalf("WriteFile(config) error = %v", err)
	}

	origTransport := http.DefaultTransport
	server.Client()
	t.Cleanup(func() {
		http.DefaultTransport = origTransport
	})

	client := server.Client()
	http.DefaultTransport = client.Transport

	if err := runCleanup(cleanupCmd, nil); err != nil {
		t.Fatalf("runCleanup() error = %v", err)
	}
	if _, err := os.Stat(statePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("state file still exists after cleanup: %v", err)
	}
}

func TestRunCleanup_RejectsEndpointMismatch(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	statePath := filepath.Join(dir, "state.yaml")
	oldConfigFile := configFile
	oldPlugin := plugin
	oldStateFile := stateFile
	t.Cleanup(func() {
		configFile = oldConfigFile
		plugin = oldPlugin
		stateFile = oldStateFile
	})

	cfgData := []byte(`sonarqube:
  endpoint: https://sonarqube-a.example.com
  manager:
    username: manager
    token: token
`)
	if err := os.WriteFile(cfgPath, cfgData, 0600); err != nil {
		t.Fatalf("WriteFile(config) error = %v", err)
	}
	if err := writeResourceState(statePath, resourceState{
		Version:            1,
		Plugin:             "tektoncd",
		Endpoint:           "https://sonarqube-b.example.com",
		TaskRunID:          "task-1",
		GroupName:          "state-group",
		UserLogin:          "state-user",
		ProjectKeys:        []string{"state-project"},
		PermissionTemplate: "state-template",
		TokenName:          "test-token-task-1",
	}); err != nil {
		t.Fatalf("writeResourceState(state) error = %v", err)
	}

	configFile = cfgPath
	stateFile = statePath
	plugin = "tektoncd"

	err := runCleanup(cleanupCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "endpoint mismatch") {
		t.Fatalf("runCleanup() error = %v, want endpoint mismatch", err)
	}
}

func TestCleanupOnError_RemovesStateFileAfterSuccessfulRollback(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "resource-state.yaml")
	tokenPath := filepath.Join(dir, "token.env")
	if err := os.WriteFile(statePath, []byte("state"), 0600); err != nil {
		t.Fatalf("WriteFile(state) error = %v", err)
	}
	if err := os.WriteFile(tokenPath, []byte("token"), 0600); err != nil {
		t.Fatalf("WriteFile(token) error = %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/authorizations/groups/test-group", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	c := sqclient.New(server.URL, "test-token")
	res := config.TempResource{
		PluginName: "tektoncd",
		Group:      config.Group{Name: "test-group"},
	}

	err := cleanupOnError(c, res, &cleanupPlan{deleteGroup: true}, "task-1", errors.New("post-create failed"), statePath, tokenPath)
	if err == nil || !strings.Contains(err.Error(), "post-create failed") {
		t.Fatalf("cleanupOnError() error = %v, want original error", err)
	}
	if _, statErr := os.Stat(statePath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("state file still exists after successful rollback: %v", statErr)
	}
	if _, statErr := os.Stat(tokenPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("token file still exists after successful rollback: %v", statErr)
	}
}
