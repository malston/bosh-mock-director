// ABOUTME: HTTP handlers for all BOSH Director API endpoints.
// ABOUTME: Implements the 16 unique endpoints needed for the 18 MCP tools.

package mockbosh

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Handlers provides HTTP handlers for the mock BOSH Director API.
type Handlers struct {
	state     *State
	simulator *TaskSimulator
	username  string
	password  string
}

// NewHandlers creates a new handlers instance.
func NewHandlers(state *State, simulator *TaskSimulator, username, password string) *Handlers {
	return &Handlers{
		state:     state,
		simulator: simulator,
		username:  username,
		password:  password,
	}
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Code:        status,
		Description: message,
	})
}

// CheckAuth validates Basic Auth credentials.
func (h *Handlers) CheckAuth(r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return user == h.username && pass == h.password
}

// HandleDeployments handles GET /deployments.
func (h *Handlers) HandleDeployments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	deployments := h.state.GetDeployments()
	writeJSON(w, http.StatusOK, deployments)
}

// HandleDeploymentVMs handles GET /deployments/:name/vms.
func (h *Handlers) HandleDeploymentVMs(w http.ResponseWriter, r *http.Request, deployment string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	vms, err := h.state.GetVMs(deployment)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, vms)
}

// HandleDeploymentInstances handles GET /deployments/:name/instances.
func (h *Handlers) HandleDeploymentInstances(w http.ResponseWriter, r *http.Request, deployment string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	instances, err := h.state.GetInstances(deployment)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	// Check if full format is requested
	format := r.URL.Query().Get("format")
	if format != "full" {
		// Strip processes for non-full format
		for i := range instances {
			instances[i].Processes = nil
		}
	}

	writeJSON(w, http.StatusOK, instances)
}

// HandleDeploymentVariables handles GET /deployments/:name/variables.
func (h *Handlers) HandleDeploymentVariables(w http.ResponseWriter, r *http.Request, deployment string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	variables, err := h.state.GetVariables(deployment)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, variables)
}

// HandleDeleteDeployment handles DELETE /deployments/:name.
func (h *Handlers) HandleDeleteDeployment(w http.ResponseWriter, r *http.Request, deployment string) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Check if deployment exists
	if !h.state.HasDeployment(deployment) {
		writeError(w, http.StatusNotFound, fmt.Sprintf("deployment '%s' not found", deployment))
		return
	}

	// Check force parameter
	force := r.URL.Query().Get("force") == "true"

	// Create task
	task := h.state.CreateTask(fmt.Sprintf("delete deployment %s", deployment), deployment, h.username)

	// Start simulation
	h.simulator.ExecuteDelete(task.ID, deployment, force)

	// Return task location
	w.Header().Set("Location", fmt.Sprintf("/tasks/%d", task.ID))
	w.WriteHeader(http.StatusFound)
}

// HandleDeploymentJobs handles PUT /deployments/:name/jobs/:job for state changes.
func (h *Handlers) HandleDeploymentJobs(w http.ResponseWriter, r *http.Request, deployment, job string) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Check if deployment exists
	if !h.state.HasDeployment(deployment) {
		writeError(w, http.StatusNotFound, fmt.Sprintf("deployment '%s' not found", deployment))
		return
	}

	// Get state parameter
	state := r.URL.Query().Get("state")
	if state == "" {
		writeError(w, http.StatusBadRequest, "state parameter is required")
		return
	}

	// Parse job and index
	jobName := job
	index := ""
	if strings.Contains(job, "/") {
		parts := strings.SplitN(job, "/", 2)
		jobName = parts[0]
		index = parts[1]
	}

	// Create task based on state
	var task *Task
	switch state {
	case "started":
		desc := fmt.Sprintf("start jobs in deployment %s", deployment)
		if jobName != "" {
			desc = fmt.Sprintf("start job %s in deployment %s", jobName, deployment)
		}
		task = h.state.CreateTask(desc, deployment, h.username)
		h.simulator.ExecuteStart(task.ID, deployment, jobName)
	case "stopped":
		desc := fmt.Sprintf("stop jobs in deployment %s", deployment)
		if jobName != "" {
			desc = fmt.Sprintf("stop job %s in deployment %s", jobName, deployment)
		}
		task = h.state.CreateTask(desc, deployment, h.username)
		h.simulator.ExecuteStop(task.ID, deployment, jobName)
	case "restart":
		desc := fmt.Sprintf("restart jobs in deployment %s", deployment)
		if jobName != "" {
			desc = fmt.Sprintf("restart job %s in deployment %s", jobName, deployment)
		}
		task = h.state.CreateTask(desc, deployment, h.username)
		h.simulator.ExecuteRestart(task.ID, deployment, jobName)
	case "recreate":
		desc := fmt.Sprintf("recreate VMs for deployment %s", deployment)
		if jobName != "" {
			desc = fmt.Sprintf("recreate VMs for %s/%s", deployment, jobName)
			if index != "" {
				desc = fmt.Sprintf("recreate VM %s/%s/%s", deployment, jobName, index)
			}
		}
		task = h.state.CreateTask(desc, deployment, h.username)
		h.simulator.ExecuteRecreate(task.ID, deployment, jobName, index)
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown state: %s", state))
		return
	}

	// Return task location
	w.Header().Set("Location", fmt.Sprintf("/tasks/%d", task.ID))
	w.WriteHeader(http.StatusFound)
}

// HandleDeploymentRecreate handles PUT /deployments/:name?state=recreate.
func (h *Handlers) HandleDeploymentRecreate(w http.ResponseWriter, r *http.Request, deployment string) {
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Check if deployment exists
	if !h.state.HasDeployment(deployment) {
		writeError(w, http.StatusNotFound, fmt.Sprintf("deployment '%s' not found", deployment))
		return
	}

	// Get state parameter
	state := r.URL.Query().Get("state")
	if state != "recreate" {
		writeError(w, http.StatusBadRequest, "state=recreate is required")
		return
	}

	// Create task
	task := h.state.CreateTask(fmt.Sprintf("recreate VMs for deployment %s", deployment), deployment, h.username)

	// Start simulation
	h.simulator.ExecuteRecreate(task.ID, deployment, "", "")

	// Return task location
	w.Header().Set("Location", fmt.Sprintf("/tasks/%d", task.ID))
	w.WriteHeader(http.StatusFound)
}

// HandleTasks handles GET /tasks.
func (h *Handlers) HandleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	state := r.URL.Query().Get("state")
	deployment := r.URL.Query().Get("deployment")
	limitStr := r.URL.Query().Get("limit")

	limit := 0
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid limit parameter")
			return
		}
	}

	tasks := h.state.GetTasks(state, deployment, limit)
	writeJSON(w, http.StatusOK, tasks)
}

// HandleTask handles GET /tasks/:id.
func (h *Handlers) HandleTask(w http.ResponseWriter, r *http.Request, taskID int) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	task, err := h.state.GetTask(taskID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, task)
}

// HandleTaskOutput handles GET /tasks/:id/output.
func (h *Handlers) HandleTaskOutput(w http.ResponseWriter, r *http.Request, taskID int) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	task, err := h.state.GetTask(taskID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	outputType := r.URL.Query().Get("type")
	output := h.simulator.GetTaskOutput(task, outputType)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}

// HandleStemcells handles GET /stemcells.
func (h *Handlers) HandleStemcells(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stemcells := h.state.GetStemcells()
	writeJSON(w, http.StatusOK, stemcells)
}

// HandleReleases handles GET /releases.
func (h *Handlers) HandleReleases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	releases := h.state.GetReleases()
	writeJSON(w, http.StatusOK, releases)
}

// HandleConfigs handles GET /configs with type and latest parameters.
func (h *Handlers) HandleConfigs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	configType := r.URL.Query().Get("type")
	// latest := r.URL.Query().Get("latest") == "true" // Not used but could filter

	switch configType {
	case "cloud":
		config := h.state.GetCloudConfig()
		if config == nil {
			writeJSON(w, http.StatusOK, []CloudConfig{})
		} else {
			writeJSON(w, http.StatusOK, []CloudConfig{*config})
		}
	case "runtime":
		configs := h.state.GetRuntimeConfigs()
		writeJSON(w, http.StatusOK, configs)
	case "cpi":
		config := h.state.GetCPIConfig()
		if config == nil {
			writeJSON(w, http.StatusOK, []CPIConfig{})
		} else {
			writeJSON(w, http.StatusOK, []CPIConfig{*config})
		}
	default:
		writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown config type: %s", configType))
	}
}

// HandleLocks handles GET /locks.
func (h *Handlers) HandleLocks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	locks := h.state.GetLocks()
	writeJSON(w, http.StatusOK, locks)
}

// HandleInfo handles GET /info for BOSH Director info.
func (h *Handlers) HandleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	info := map[string]interface{}{
		"name":         "Mock BOSH Director",
		"uuid":         "mock-bosh-director-uuid",
		"version":      "281.0.0 (00000000)",
		"user":         h.username,
		"cpi":          "google_cpi",
		"stemcell_os":  "ubuntu-jammy",
		"user_authentication": map[string]interface{}{
			"type": "basic",
		},
	}
	writeJSON(w, http.StatusOK, info)
}
