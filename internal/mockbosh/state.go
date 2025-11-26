// ABOUTME: Thread-safe state manager for the mock BOSH Director.
// ABOUTME: Provides CRUD operations for all BOSH resources with mutex protection.

package mockbosh

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// StateData holds all mock BOSH Director data.
type StateData struct {
	mu             sync.RWMutex
	Deployments    map[string]*Deployment
	VMs            map[string][]VM
	Instances      map[string][]Instance
	Variables      map[string][]Variable
	Tasks          map[int]*Task
	Stemcells      []Stemcell
	Releases       []Release
	CloudConfig    *CloudConfig
	RuntimeConfigs []RuntimeConfig
	CPIConfig      *CPIConfig
	Locks          []Lock
	nextTaskID     int
}

// State wraps StateData with thread-safe operations.
type State struct {
	data *StateData
}

// NewState creates a new state manager with default fixtures.
func NewState() *State {
	return &State{data: DefaultFixtures()}
}

// NewStateWithData creates a new state manager with custom data.
func NewStateWithData(data *StateData) *State {
	return &State{data: data}
}

// GetDeployments returns all deployments.
func (s *State) GetDeployments() []Deployment {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	result := make([]Deployment, 0, len(s.data.Deployments))
	for _, d := range s.data.Deployments {
		result = append(result, *d)
	}
	return result
}

// GetDeployment returns a deployment by name.
func (s *State) GetDeployment(name string) (*Deployment, error) {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	d, ok := s.data.Deployments[name]
	if !ok {
		return nil, fmt.Errorf("deployment '%s' not found", name)
	}
	copy := *d
	return &copy, nil
}

// DeleteDeployment removes a deployment and associated resources.
func (s *State) DeleteDeployment(name string) error {
	s.data.mu.Lock()
	defer s.data.mu.Unlock()

	if _, ok := s.data.Deployments[name]; !ok {
		return fmt.Errorf("deployment '%s' not found", name)
	}

	delete(s.data.Deployments, name)
	delete(s.data.VMs, name)
	delete(s.data.Instances, name)
	delete(s.data.Variables, name)

	// Update stemcell deployment references
	for i := range s.data.Stemcells {
		deps := make([]string, 0)
		for _, d := range s.data.Stemcells[i].Deployments {
			if d != name {
				deps = append(deps, d)
			}
		}
		s.data.Stemcells[i].Deployments = deps
	}

	return nil
}

// GetVMs returns VMs for a deployment.
func (s *State) GetVMs(deployment string) ([]VM, error) {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	if _, ok := s.data.Deployments[deployment]; !ok {
		return nil, fmt.Errorf("deployment '%s' not found", deployment)
	}

	vms := s.data.VMs[deployment]
	result := make([]VM, len(vms))
	copy(result, vms)
	return result, nil
}

// GetInstances returns instances for a deployment.
func (s *State) GetInstances(deployment string) ([]Instance, error) {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	if _, ok := s.data.Deployments[deployment]; !ok {
		return nil, fmt.Errorf("deployment '%s' not found", deployment)
	}

	instances := s.data.Instances[deployment]
	result := make([]Instance, len(instances))
	copy(result, instances)
	return result, nil
}

// GetVariables returns variables for a deployment.
func (s *State) GetVariables(deployment string) ([]Variable, error) {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	if _, ok := s.data.Deployments[deployment]; !ok {
		return nil, fmt.Errorf("deployment '%s' not found", deployment)
	}

	vars := s.data.Variables[deployment]
	result := make([]Variable, len(vars))
	copy(result, vars)
	return result, nil
}

// GetTasks returns tasks matching the filter.
func (s *State) GetTasks(state, deployment string, limit int) []Task {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	result := make([]Task, 0)
	for _, t := range s.data.Tasks {
		if state != "" && t.State != state {
			continue
		}
		if deployment != "" && t.Deployment != deployment {
			continue
		}
		result = append(result, *t)
	}

	// Sort by ID descending (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID > result[j].ID
	})

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

// GetTask returns a task by ID.
func (s *State) GetTask(id int) (*Task, error) {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	t, ok := s.data.Tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %d not found", id)
	}
	copy := *t
	return &copy, nil
}

// CreateTask creates a new task and returns its ID.
func (s *State) CreateTask(description, deployment, user string) *Task {
	s.data.mu.Lock()
	defer s.data.mu.Unlock()

	s.data.nextTaskID++
	task := &Task{
		ID:          s.data.nextTaskID,
		State:       "queued",
		Description: description,
		Timestamp:   time.Now().Unix(),
		User:        user,
		Deployment:  deployment,
	}
	s.data.Tasks[task.ID] = task
	return task
}

// UpdateTaskState updates a task's state.
func (s *State) UpdateTaskState(id int, state, result string) error {
	s.data.mu.Lock()
	defer s.data.mu.Unlock()

	t, ok := s.data.Tasks[id]
	if !ok {
		return fmt.Errorf("task %d not found", id)
	}
	t.State = state
	if result != "" {
		t.Result = result
	}
	return nil
}

// GetStemcells returns all stemcells.
func (s *State) GetStemcells() []Stemcell {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	result := make([]Stemcell, len(s.data.Stemcells))
	copy(result, s.data.Stemcells)
	return result
}

// GetReleases returns all releases.
func (s *State) GetReleases() []Release {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	result := make([]Release, len(s.data.Releases))
	copy(result, s.data.Releases)
	return result
}

// GetCloudConfig returns the cloud config.
func (s *State) GetCloudConfig() *CloudConfig {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	if s.data.CloudConfig == nil {
		return nil
	}
	copy := *s.data.CloudConfig
	return &copy
}

// GetRuntimeConfigs returns all runtime configs.
func (s *State) GetRuntimeConfigs() []RuntimeConfig {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	result := make([]RuntimeConfig, len(s.data.RuntimeConfigs))
	copy(result, s.data.RuntimeConfigs)
	return result
}

// GetCPIConfig returns the CPI config.
func (s *State) GetCPIConfig() *CPIConfig {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	if s.data.CPIConfig == nil {
		return nil
	}
	copy := *s.data.CPIConfig
	return &copy
}

// GetLocks returns all locks.
func (s *State) GetLocks() []Lock {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()

	result := make([]Lock, len(s.data.Locks))
	copy(result, s.data.Locks)
	return result
}

// AddLock adds a deployment lock.
func (s *State) AddLock(lockType, resource, taskID string, timeout time.Duration) {
	s.data.mu.Lock()
	defer s.data.mu.Unlock()

	s.data.Locks = append(s.data.Locks, Lock{
		Type:     lockType,
		Resource: resource,
		Timeout:  timeout.String(),
		TaskID:   taskID,
	})
}

// RemoveLock removes a lock for a resource.
func (s *State) RemoveLock(resource string) {
	s.data.mu.Lock()
	defer s.data.mu.Unlock()

	locks := make([]Lock, 0)
	for _, l := range s.data.Locks {
		if l.Resource != resource {
			locks = append(locks, l)
		}
	}
	s.data.Locks = locks
}

// RecreateVMs marks VMs as recreating and updates their state.
func (s *State) RecreateVMs(deployment, job, index string) error {
	s.data.mu.Lock()
	defer s.data.mu.Unlock()

	if _, ok := s.data.Deployments[deployment]; !ok {
		return fmt.Errorf("deployment '%s' not found", deployment)
	}

	// Update VMs
	vms := s.data.VMs[deployment]
	for i := range vms {
		if job != "" && vms[i].Job != job {
			continue
		}
		if index != "" && fmt.Sprintf("%d", vms[i].Index) != index {
			continue
		}
		// Simulate recreation by generating new VM CID
		vms[i].VMCID = fmt.Sprintf("vm-%s-%s-%d-recreated", deployment, vms[i].Job, vms[i].Index)
	}

	return nil
}

// ChangeJobState changes the state of jobs in a deployment.
func (s *State) ChangeJobState(deployment, job, newState string) error {
	s.data.mu.Lock()
	defer s.data.mu.Unlock()

	if _, ok := s.data.Deployments[deployment]; !ok {
		return fmt.Errorf("deployment '%s' not found", deployment)
	}

	// Determine process state based on job state
	processState := "running"
	vmProcessState := "running"
	switch newState {
	case "stopped":
		processState = "stopped"
		vmProcessState = "stopped"
	case "started", "restart":
		processState = "running"
		vmProcessState = "running"
	}

	// Update VMs
	vms := s.data.VMs[deployment]
	for i := range vms {
		if job != "" && vms[i].Job != job {
			continue
		}
		vms[i].ProcessState = vmProcessState
		if newState == "stopped" {
			vms[i].State = "stopped"
		} else {
			vms[i].State = "started"
		}
	}

	// Update instances and their processes
	instances := s.data.Instances[deployment]
	for i := range instances {
		if job != "" && instances[i].Job != job {
			continue
		}
		instances[i].State = processState
		for j := range instances[i].Processes {
			instances[i].Processes[j].State = processState
		}
	}

	return nil
}

// HasDeployment checks if a deployment exists.
func (s *State) HasDeployment(name string) bool {
	s.data.mu.RLock()
	defer s.data.mu.RUnlock()
	_, ok := s.data.Deployments[name]
	return ok
}
