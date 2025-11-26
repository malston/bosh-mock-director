// ABOUTME: Tests for HTTP handlers.
// ABOUTME: Verifies API endpoint responses and error handling.

package mockbosh

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestHandlers() *Handlers {
	state := NewState()
	simulator := NewTaskSimulator(state, 10.0, false) // Fast simulation
	return NewHandlers(state, simulator, "admin", "admin")
}

func TestHandleDeployments(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/deployments", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleDeployments(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var deployments []Deployment
	if err := json.Unmarshal(w.Body.Bytes(), &deployments); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(deployments) == 0 {
		t.Error("Expected deployments in response")
	}
}

func TestHandleDeploymentVMs(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/deployments/cf/vms", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleDeploymentVMs(w, req, "cf")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var vms []VM
	if err := json.Unmarshal(w.Body.Bytes(), &vms); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(vms) == 0 {
		t.Error("Expected VMs in response")
	}
}

func TestHandleDeploymentVMsNotFound(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/deployments/nonexistent/vms", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleDeploymentVMs(w, req, "nonexistent")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleDeploymentInstances(t *testing.T) {
	handlers := setupTestHandlers()

	// With format=full
	req := httptest.NewRequest(http.MethodGet, "/deployments/cf/instances?format=full", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleDeploymentInstances(w, req, "cf")

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var instances []Instance
	if err := json.Unmarshal(w.Body.Bytes(), &instances); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(instances) == 0 {
		t.Error("Expected instances in response")
	}

	// Check that processes are included with format=full
	hasProcesses := false
	for _, inst := range instances {
		if len(inst.Processes) > 0 {
			hasProcesses = true
			break
		}
	}
	if !hasProcesses {
		t.Error("Expected processes with format=full")
	}
}

func TestHandleTasks(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var tasks []Task
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(tasks) == 0 {
		t.Error("Expected tasks in response")
	}
}

func TestHandleTasksWithFilters(t *testing.T) {
	handlers := setupTestHandlers()

	// Filter by state
	req := httptest.NewRequest(http.MethodGet, "/tasks?state=done", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var tasks []Task
	if err := json.Unmarshal(w.Body.Bytes(), &tasks); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	for _, task := range tasks {
		if task.State != "done" {
			t.Errorf("Expected all tasks to have state 'done', got '%s'", task.State)
		}
	}
}

func TestHandleTask(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleTask(w, req, 1)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var task Task
	if err := json.Unmarshal(w.Body.Bytes(), &task); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if task.ID != 1 {
		t.Errorf("Expected task ID 1, got %d", task.ID)
	}
}

func TestHandleTaskNotFound(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/tasks/99999", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleTask(w, req, 99999)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestHandleStemcells(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/stemcells", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleStemcells(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var stemcells []Stemcell
	if err := json.Unmarshal(w.Body.Bytes(), &stemcells); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(stemcells) == 0 {
		t.Error("Expected stemcells in response")
	}
}

func TestHandleReleases(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/releases", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleReleases(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var releases []Release
	if err := json.Unmarshal(w.Body.Bytes(), &releases); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(releases) == 0 {
		t.Error("Expected releases in response")
	}
}

func TestHandleConfigs(t *testing.T) {
	handlers := setupTestHandlers()

	testCases := []struct {
		configType string
		expectLen  int
	}{
		{"cloud", 1},
		{"runtime", 2},
		{"cpi", 1},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/configs?type="+tc.configType+"&latest=true", nil)
		req.SetBasicAuth("admin", "admin")
		w := httptest.NewRecorder()

		handlers.HandleConfigs(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d for %s config, got %d", http.StatusOK, tc.configType, w.Code)
		}
	}
}

func TestHandleLocks(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/locks", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleLocks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var locks []Lock
	if err := json.Unmarshal(w.Body.Bytes(), &locks); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
}

func TestHandleInfo(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	w := httptest.NewRecorder()

	handlers.HandleInfo(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var info map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if info["name"] != "Mock BOSH Director" {
		t.Errorf("Expected name 'Mock BOSH Director', got '%s'", info["name"])
	}
}

func TestHandleDeleteDeployment(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodDelete, "/deployments/redis", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleDeleteDeployment(w, req, "redis")

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Error("Expected Location header")
	}
}

func TestHandleDeleteDeploymentNotFound(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodDelete, "/deployments/nonexistent", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleDeleteDeployment(w, req, "nonexistent")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCheckAuth(t *testing.T) {
	handlers := setupTestHandlers()

	// Valid auth
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("admin", "admin")
	if !handlers.CheckAuth(req) {
		t.Error("Expected valid auth to pass")
	}

	// Invalid password
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth("admin", "wrong")
	if handlers.CheckAuth(req) {
		t.Error("Expected invalid auth to fail")
	}

	// No auth
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	if handlers.CheckAuth(req) {
		t.Error("Expected missing auth to fail")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	handlers := setupTestHandlers()

	req := httptest.NewRequest(http.MethodPost, "/deployments", nil)
	req.SetBasicAuth("admin", "admin")
	w := httptest.NewRecorder()

	handlers.HandleDeployments(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}
