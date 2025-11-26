// ABOUTME: Simulates BOSH task execution and progression.
// ABOUTME: Runs background goroutines that transition tasks through states.

package mockbosh

import (
	"fmt"
	"log"
	"time"
)

// TaskSimulator manages task execution simulation.
type TaskSimulator struct {
	state *State
	speed float64 // Simulation speed multiplier (1.0 = normal, 10.0 = 10x faster)
	debug bool
}

// NewTaskSimulator creates a new task simulator.
func NewTaskSimulator(state *State, speed float64, debug bool) *TaskSimulator {
	if speed <= 0 {
		speed = 1.0
	}
	return &TaskSimulator{
		state: state,
		speed: speed,
		debug: debug,
	}
}

// scaledDuration returns a duration scaled by the simulation speed.
func (ts *TaskSimulator) scaledDuration(d time.Duration) time.Duration {
	return time.Duration(float64(d) / ts.speed)
}

// log prints debug messages if debug mode is enabled.
func (ts *TaskSimulator) log(format string, args ...interface{}) {
	if ts.debug {
		log.Printf("[TaskSimulator] "+format, args...)
	}
}

// ExecuteDelete simulates a deployment deletion.
func (ts *TaskSimulator) ExecuteDelete(taskID int, deployment string, force bool) {
	go func() {
		ts.log("Task %d: Starting delete deployment %s (force=%v)", taskID, deployment, force)

		// Queue → Processing
		time.Sleep(ts.scaledDuration(500 * time.Millisecond))
		ts.state.UpdateTaskState(taskID, "processing", "")
		ts.log("Task %d: Processing", taskID)

		// Add lock
		ts.state.AddLock("deployment", deployment, fmt.Sprintf("%d", taskID), 30*time.Minute)

		// Simulate deletion work
		time.Sleep(ts.scaledDuration(2 * time.Second))

		// Perform deletion
		err := ts.state.DeleteDeployment(deployment)
		if err != nil {
			ts.state.UpdateTaskState(taskID, "error", err.Error())
			ts.log("Task %d: Error - %s", taskID, err.Error())
			ts.state.RemoveLock(deployment)
			return
		}

		// Remove lock and complete
		ts.state.RemoveLock(deployment)
		ts.state.UpdateTaskState(taskID, "done", fmt.Sprintf("Deleted deployment %s", deployment))
		ts.log("Task %d: Done", taskID)
	}()
}

// ExecuteRecreate simulates VM recreation.
func (ts *TaskSimulator) ExecuteRecreate(taskID int, deployment, job, index string) {
	go func() {
		ts.log("Task %d: Starting recreate %s/%s/%s", taskID, deployment, job, index)

		// Queue → Processing
		time.Sleep(ts.scaledDuration(500 * time.Millisecond))
		ts.state.UpdateTaskState(taskID, "processing", "")
		ts.log("Task %d: Processing", taskID)

		// Add lock
		ts.state.AddLock("deployment", deployment, fmt.Sprintf("%d", taskID), 30*time.Minute)

		// Simulate recreation work (longer for recreate)
		time.Sleep(ts.scaledDuration(3 * time.Second))

		// Perform recreation
		err := ts.state.RecreateVMs(deployment, job, index)
		if err != nil {
			ts.state.UpdateTaskState(taskID, "error", err.Error())
			ts.log("Task %d: Error - %s", taskID, err.Error())
			ts.state.RemoveLock(deployment)
			return
		}

		// Remove lock and complete
		ts.state.RemoveLock(deployment)

		result := fmt.Sprintf("Recreated VMs for deployment %s", deployment)
		if job != "" {
			result = fmt.Sprintf("Recreated VMs for %s/%s", deployment, job)
			if index != "" {
				result = fmt.Sprintf("Recreated VM %s/%s/%s", deployment, job, index)
			}
		}
		ts.state.UpdateTaskState(taskID, "done", result)
		ts.log("Task %d: Done", taskID)
	}()
}

// ExecuteStart simulates starting jobs.
func (ts *TaskSimulator) ExecuteStart(taskID int, deployment, job string) {
	go func() {
		ts.log("Task %d: Starting start %s/%s", taskID, deployment, job)

		// Queue → Processing
		time.Sleep(ts.scaledDuration(500 * time.Millisecond))
		ts.state.UpdateTaskState(taskID, "processing", "")
		ts.log("Task %d: Processing", taskID)

		// Add lock
		ts.state.AddLock("deployment", deployment, fmt.Sprintf("%d", taskID), 30*time.Minute)

		// Simulate start work
		time.Sleep(ts.scaledDuration(1 * time.Second))

		// Perform state change
		err := ts.state.ChangeJobState(deployment, job, "started")
		if err != nil {
			ts.state.UpdateTaskState(taskID, "error", err.Error())
			ts.log("Task %d: Error - %s", taskID, err.Error())
			ts.state.RemoveLock(deployment)
			return
		}

		// Remove lock and complete
		ts.state.RemoveLock(deployment)

		result := fmt.Sprintf("Started jobs in deployment %s", deployment)
		if job != "" {
			result = fmt.Sprintf("Started job %s in deployment %s", job, deployment)
		}
		ts.state.UpdateTaskState(taskID, "done", result)
		ts.log("Task %d: Done", taskID)
	}()
}

// ExecuteStop simulates stopping jobs.
func (ts *TaskSimulator) ExecuteStop(taskID int, deployment, job string) {
	go func() {
		ts.log("Task %d: Starting stop %s/%s", taskID, deployment, job)

		// Queue → Processing
		time.Sleep(ts.scaledDuration(500 * time.Millisecond))
		ts.state.UpdateTaskState(taskID, "processing", "")
		ts.log("Task %d: Processing", taskID)

		// Add lock
		ts.state.AddLock("deployment", deployment, fmt.Sprintf("%d", taskID), 30*time.Minute)

		// Simulate stop work
		time.Sleep(ts.scaledDuration(1 * time.Second))

		// Perform state change
		err := ts.state.ChangeJobState(deployment, job, "stopped")
		if err != nil {
			ts.state.UpdateTaskState(taskID, "error", err.Error())
			ts.log("Task %d: Error - %s", taskID, err.Error())
			ts.state.RemoveLock(deployment)
			return
		}

		// Remove lock and complete
		ts.state.RemoveLock(deployment)

		result := fmt.Sprintf("Stopped jobs in deployment %s", deployment)
		if job != "" {
			result = fmt.Sprintf("Stopped job %s in deployment %s", job, deployment)
		}
		ts.state.UpdateTaskState(taskID, "done", result)
		ts.log("Task %d: Done", taskID)
	}()
}

// ExecuteRestart simulates restarting jobs.
func (ts *TaskSimulator) ExecuteRestart(taskID int, deployment, job string) {
	go func() {
		ts.log("Task %d: Starting restart %s/%s", taskID, deployment, job)

		// Queue → Processing
		time.Sleep(ts.scaledDuration(500 * time.Millisecond))
		ts.state.UpdateTaskState(taskID, "processing", "")
		ts.log("Task %d: Processing", taskID)

		// Add lock
		ts.state.AddLock("deployment", deployment, fmt.Sprintf("%d", taskID), 30*time.Minute)

		// Simulate stop
		time.Sleep(ts.scaledDuration(1 * time.Second))
		if err := ts.state.ChangeJobState(deployment, job, "stopped"); err != nil {
			ts.state.UpdateTaskState(taskID, "error", err.Error())
			ts.log("Task %d: Error - %s", taskID, err.Error())
			ts.state.RemoveLock(deployment)
			return
		}

		// Simulate start
		time.Sleep(ts.scaledDuration(1 * time.Second))
		err := ts.state.ChangeJobState(deployment, job, "started")
		if err != nil {
			ts.state.UpdateTaskState(taskID, "error", err.Error())
			ts.log("Task %d: Error - %s", taskID, err.Error())
			ts.state.RemoveLock(deployment)
			return
		}

		// Remove lock and complete
		ts.state.RemoveLock(deployment)

		result := fmt.Sprintf("Restarted jobs in deployment %s", deployment)
		if job != "" {
			result = fmt.Sprintf("Restarted job %s in deployment %s", job, deployment)
		}
		ts.state.UpdateTaskState(taskID, "done", result)
		ts.log("Task %d: Done", taskID)
	}()
}

// GetTaskOutput returns simulated task output.
func (ts *TaskSimulator) GetTaskOutput(task *Task, outputType string) string {
	if outputType == "" {
		outputType = "result"
	}

	switch outputType {
	case "result":
		if task.Result != "" {
			return task.Result
		}
		return fmt.Sprintf("Task %d: %s", task.ID, task.Description)
	case "debug":
		return fmt.Sprintf("DEBUG: Task %d started at %d\nDEBUG: State: %s\nDEBUG: Deployment: %s",
			task.ID, task.Timestamp, task.State, task.Deployment)
	case "cpi":
		return fmt.Sprintf("CPI: No CPI operations for task %d", task.ID)
	case "event":
		return fmt.Sprintf("EVENT: Task %d %s at %d", task.ID, task.State, task.Timestamp)
	default:
		return task.Result
	}
}
