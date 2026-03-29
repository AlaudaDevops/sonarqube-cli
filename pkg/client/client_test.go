package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tektoncd/operator/tools/sonarqube-cli/pkg/config"
)

func TestClient_CreateReturnsAlreadyExists(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	c := New(server.URL, "test-token")

	mux.HandleFunc("/api/projects/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"errors":[{"msg":"Project already exists"}]}`)
	})

	created, err := c.CreateProject(config.Project{Key: "test", Name: "test"})
	if created {
		t.Fatal("CreateProject() created = true, want false")
	}
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("CreateProject() error = %v, want ErrAlreadyExists", err)
	}
}

func TestClient_AddGlobalPermissionReturnsAlreadyExists(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	c := New(server.URL, "test-token")

	mux.HandleFunc("/api/permissions/add_user", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"errors":[{"msg":"This permission has already been granted to this user"}]}`)
	})

	err := c.AddGlobalPermission("user", "scan")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("AddGlobalPermission() error = %v, want ErrAlreadyExists", err)
	}
}

func TestClient_DeleteProjectIgnores404(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	c := New(server.URL, "test-token")

	mux.HandleFunc("/api/projects/delete", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"errors":[{"msg":"Project not found"}]}`)
	})

	if err := c.DeleteProject("test"); err != nil {
		t.Fatalf("DeleteProject() error = %v, want nil", err)
	}
}

func TestClient_CreatePermissionTemplateReturnsAlreadyExists(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	c := New(server.URL, "test-token")

	mux.HandleFunc("/api/permissions/create_template", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"errors":[{"msg":"Permission template already exists"}]}`)
	})

	tmpl := config.PermissionTemplate{
		Name:        "test-tmpl",
		Permissions: []string{"user", "scan"},
	}

	created, err := c.CreatePermissionTemplate(tmpl, "test-group")
	if created {
		t.Fatal("CreatePermissionTemplate() created = true, want false")
	}
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("CreatePermissionTemplate() error = %v, want ErrAlreadyExists", err)
	}
}

func TestClient_GenerateUserTokenDoesNotRevokeExistingToken(t *testing.T) {
	var revokeCalls int

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	c := New(server.URL, "test-token")

	mux.HandleFunc("/api/user_tokens/revoke", func(w http.ResponseWriter, r *http.Request) {
		revokeCalls++
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/user_tokens/generate", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"token":"generated-token"}`)
	})

	token, err := c.GenerateUserToken("user", "token-name")
	if err != nil {
		t.Fatalf("GenerateUserToken() error = %v", err)
	}
	if token != "generated-token" {
		t.Fatalf("GenerateUserToken() token = %q, want generated-token", token)
	}
	if revokeCalls != 0 {
		t.Fatalf("GenerateUserToken() revokeCalls = %d, want 0", revokeCalls)
	}
}
