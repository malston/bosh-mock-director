// ABOUTME: Tests for the thread-safe state manager.
// ABOUTME: Verifies CRUD operations and concurrent access.

package mockbosh

import (
	"sync"
	"testing"
)

func TestNewState(t *testing.T) {
	state := NewState()
	if state == nil {
		t.Fatal("NewState returned nil")
	}

	deployments := state.GetDeployments()
	if len(deployments) == 0 {
		t.Error("Expected default deployments, got none")
	}
}

func TestGetDeployments(t *testing.T) {
	state := NewState()
	deployments := state.GetDeployments()

	// Default fixtures include cf, redis, mysql
	names := make(map[string]bool)
	for _, d := range deployments {
		names[d.Name] = true
	}

	if !names["cf"] {
		t.Error("Expected 'cf' deployment")
	}
	if !names["redis"] {
		t.Error("Expected 'redis' deployment")
	}
	if !names["mysql"] {
		t.Error("Expected 'mysql' deployment")
	}
}

func TestGetDeployment(t *testing.T) {
	state := NewState()

	d, err := state.GetDeployment("cf")
	if err != nil {
		t.Fatalf("GetDeployment failed: %v", err)
	}
	if d.Name != "cf" {
		t.Errorf("Expected name 'cf', got '%s'", d.Name)
	}

	_, err = state.GetDeployment("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent deployment")
	}
}

func TestDeleteDeployment(t *testing.T) {
	state := NewState()

	// Verify deployment exists
	_, err := state.GetDeployment("redis")
	if err != nil {
		t.Fatalf("redis deployment should exist: %v", err)
	}

	// Delete it
	err = state.DeleteDeployment("redis")
	if err != nil {
		t.Fatalf("DeleteDeployment failed: %v", err)
	}

	// Verify it's gone
	_, err = state.GetDeployment("redis")
	if err == nil {
		t.Error("redis deployment should be deleted")
	}

	// Delete nonexistent
	err = state.DeleteDeployment("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent deployment")
	}
}

func TestGetVMs(t *testing.T) {
	state := NewState()

	vms, err := state.GetVMs("cf")
	if err != nil {
		t.Fatalf("GetVMs failed: %v", err)
	}
	if len(vms) == 0 {
		t.Error("Expected VMs for cf deployment")
	}

	// Check VM properties
	found := false
	for _, vm := range vms {
		if vm.Job == "diego_cell" && vm.Index == 0 {
			found = true
			if vm.ProcessState != "running" {
				t.Errorf("Expected process_state 'running', got '%s'", vm.ProcessState)
			}
		}
	}
	if !found {
		t.Error("Expected diego_cell/0 VM")
	}

	_, err = state.GetVMs("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent deployment")
	}
}

func TestGetInstances(t *testing.T) {
	state := NewState()

	instances, err := state.GetInstances("cf")
	if err != nil {
		t.Fatalf("GetInstances failed: %v", err)
	}
	if len(instances) == 0 {
		t.Error("Expected instances for cf deployment")
	}

	// Check instance has processes
	found := false
	for _, inst := range instances {
		if inst.Job == "diego_cell" && inst.Index == 0 {
			found = true
			if len(inst.Processes) == 0 {
				t.Error("Expected processes for diego_cell/0")
			}
		}
	}
	if !found {
		t.Error("Expected diego_cell/0 instance")
	}
}

func TestGetTasks(t *testing.T) {
	state := NewState()

	// Get all tasks
	tasks := state.GetTasks("", "", 0)
	if len(tasks) == 0 {
		t.Error("Expected default tasks")
	}

	// Filter by state
	doneTasks := state.GetTasks("done", "", 0)
	for _, task := range doneTasks {
		if task.State != "done" {
			t.Errorf("Expected state 'done', got '%s'", task.State)
		}
	}

	// Filter by deployment
	cfTasks := state.GetTasks("", "cf", 0)
	for _, task := range cfTasks {
		if task.Deployment != "cf" {
			t.Errorf("Expected deployment 'cf', got '%s'", task.Deployment)
		}
	}

	// Limit
	limitedTasks := state.GetTasks("", "", 2)
	if len(limitedTasks) > 2 {
		t.Errorf("Expected at most 2 tasks, got %d", len(limitedTasks))
	}
}

func TestCreateTask(t *testing.T) {
	state := NewState()

	task := state.CreateTask("test task", "cf", "admin")
	if task.ID == 0 {
		t.Error("Expected non-zero task ID")
	}
	if task.State != "queued" {
		t.Errorf("Expected state 'queued', got '%s'", task.State)
	}
	if task.Description != "test task" {
		t.Errorf("Expected description 'test task', got '%s'", task.Description)
	}

	// Create another and verify ID increments
	task2 := state.CreateTask("test task 2", "cf", "admin")
	if task2.ID <= task.ID {
		t.Error("Expected task ID to increment")
	}
}

func TestUpdateTaskState(t *testing.T) {
	state := NewState()

	task := state.CreateTask("test task", "cf", "admin")

	err := state.UpdateTaskState(task.ID, "processing", "")
	if err != nil {
		t.Fatalf("UpdateTaskState failed: %v", err)
	}

	updated, _ := state.GetTask(task.ID)
	if updated.State != "processing" {
		t.Errorf("Expected state 'processing', got '%s'", updated.State)
	}

	err = state.UpdateTaskState(task.ID, "done", "completed")
	if err != nil {
		t.Fatalf("UpdateTaskState failed: %v", err)
	}

	updated, _ = state.GetTask(task.ID)
	if updated.State != "done" || updated.Result != "completed" {
		t.Errorf("Expected state 'done' and result 'completed', got '%s' and '%s'", updated.State, updated.Result)
	}

	err = state.UpdateTaskState(99999, "done", "")
	if err == nil {
		t.Error("Expected error for nonexistent task")
	}
}

func TestGetStemcells(t *testing.T) {
	state := NewState()

	stemcells := state.GetStemcells()
	if len(stemcells) == 0 {
		t.Error("Expected default stemcells")
	}

	found := false
	for _, s := range stemcells {
		if s.OperatingSystem == "ubuntu-jammy" {
			found = true
		}
	}
	if !found {
		t.Error("Expected ubuntu-jammy stemcell")
	}
}

func TestGetReleases(t *testing.T) {
	state := NewState()

	releases := state.GetReleases()
	if len(releases) == 0 {
		t.Error("Expected default releases")
	}

	found := false
	for _, r := range releases {
		if r.Name == "cf-deployment" {
			found = true
		}
	}
	if !found {
		t.Error("Expected cf-deployment release")
	}
}

func TestGetConfigs(t *testing.T) {
	state := NewState()

	cloudConfig := state.GetCloudConfig()
	if cloudConfig == nil {
		t.Error("Expected cloud config")
	}

	runtimeConfigs := state.GetRuntimeConfigs()
	if len(runtimeConfigs) == 0 {
		t.Error("Expected runtime configs")
	}

	cpiConfig := state.GetCPIConfig()
	if cpiConfig == nil {
		t.Error("Expected CPI config")
	}
}

func TestChangeJobState(t *testing.T) {
	state := NewState()

	// Stop jobs
	err := state.ChangeJobState("cf", "router", "stopped")
	if err != nil {
		t.Fatalf("ChangeJobState failed: %v", err)
	}

	vms, _ := state.GetVMs("cf")
	for _, vm := range vms {
		if vm.Job == "router" {
			if vm.ProcessState != "stopped" {
				t.Errorf("Expected process_state 'stopped', got '%s'", vm.ProcessState)
			}
		}
	}

	// Start jobs
	err = state.ChangeJobState("cf", "router", "started")
	if err != nil {
		t.Fatalf("ChangeJobState failed: %v", err)
	}

	vms, _ = state.GetVMs("cf")
	for _, vm := range vms {
		if vm.Job == "router" {
			if vm.ProcessState != "running" {
				t.Errorf("Expected process_state 'running', got '%s'", vm.ProcessState)
			}
		}
	}

	err = state.ChangeJobState("nonexistent", "", "stopped")
	if err == nil {
		t.Error("Expected error for nonexistent deployment")
	}
}

func TestConcurrentAccess(t *testing.T) {
	state := NewState()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			state.GetDeployments()
			state.GetVMs("cf")
			state.GetTasks("", "", 0)
			state.CreateTask("concurrent test", "cf", "admin")
		}()
	}
	wg.Wait()
}
