// ABOUTME: Provides default sample data for the mock BOSH Director.
// ABOUTME: Creates realistic deployments, VMs, tasks, stemcells, and releases.

package mockbosh

import (
	"fmt"
	"time"
)

// DefaultFixtures returns a fully populated set of sample data.
func DefaultFixtures() *StateData {
	now := time.Now()

	return &StateData{
		Deployments: defaultDeployments(),
		VMs:         defaultVMs(),
		Instances:   defaultInstances(),
		Variables:   defaultVariables(),
		Tasks:       defaultTasks(now),
		Stemcells:   defaultStemcells(),
		Releases:    defaultReleases(),
		CloudConfig: defaultCloudConfig(now),
		RuntimeConfigs: defaultRuntimeConfigs(now),
		CPIConfig:   defaultCPIConfig(now),
		Locks:       []Lock{},
		nextTaskID:  100,
	}
}

func defaultDeployments() map[string]*Deployment {
	return map[string]*Deployment{
		"cf": {
			Name:        "cf",
			CloudConfig: "latest",
			Releases: []NameVersion{
				{Name: "cf-deployment", Version: "40.0.0"},
				{Name: "cflinuxfs4", Version: "1.50.0"},
				{Name: "diego", Version: "2.80.0"},
				{Name: "garden-runc", Version: "1.28.0"},
			},
			Stemcells: []NameVersion{
				{Name: "bosh-google-kvm-ubuntu-jammy-go_agent", Version: "1.200"},
			},
		},
		"redis": {
			Name:        "redis",
			CloudConfig: "latest",
			Releases: []NameVersion{
				{Name: "redis", Version: "16.0.0"},
			},
			Stemcells: []NameVersion{
				{Name: "bosh-google-kvm-ubuntu-jammy-go_agent", Version: "1.200"},
			},
		},
		"mysql": {
			Name:        "mysql",
			CloudConfig: "latest",
			Releases: []NameVersion{
				{Name: "pxc", Version: "0.42.0"},
			},
			Stemcells: []NameVersion{
				{Name: "bosh-google-kvm-ubuntu-jammy-go_agent", Version: "1.200"},
			},
		},
	}
}

func defaultVMs() map[string][]VM {
	return map[string][]VM{
		"cf": {
			{
				VMCID: "vm-cf-diego-cell-0", Active: true, AgentID: "agent-cf-dc0",
				AZ: "z1", Bootstrap: false, Deployment: "cf", IPs: []string{"10.0.1.10"},
				Job: "diego_cell", Index: 0, ID: "cf-dc0-id", ProcessState: "running",
				State: "started", VMType: "large", Ignore: false,
			},
			{
				VMCID: "vm-cf-diego-cell-1", Active: true, AgentID: "agent-cf-dc1",
				AZ: "z2", Bootstrap: false, Deployment: "cf", IPs: []string{"10.0.2.10"},
				Job: "diego_cell", Index: 1, ID: "cf-dc1-id", ProcessState: "running",
				State: "started", VMType: "large", Ignore: false,
			},
			{
				VMCID: "vm-cf-diego-cell-2", Active: true, AgentID: "agent-cf-dc2",
				AZ: "z3", Bootstrap: false, Deployment: "cf", IPs: []string{"10.0.3.10"},
				Job: "diego_cell", Index: 2, ID: "cf-dc2-id", ProcessState: "running",
				State: "started", VMType: "large", Ignore: false,
			},
			{
				VMCID: "vm-cf-router-0", Active: true, AgentID: "agent-cf-r0",
				AZ: "z1", Bootstrap: true, Deployment: "cf", IPs: []string{"10.0.1.20"},
				Job: "router", Index: 0, ID: "cf-r0-id", ProcessState: "running",
				State: "started", VMType: "medium", Ignore: false,
			},
			{
				VMCID: "vm-cf-router-1", Active: true, AgentID: "agent-cf-r1",
				AZ: "z2", Bootstrap: false, Deployment: "cf", IPs: []string{"10.0.2.20"},
				Job: "router", Index: 1, ID: "cf-r1-id", ProcessState: "running",
				State: "started", VMType: "medium", Ignore: false,
			},
			{
				VMCID: "vm-cf-api-0", Active: true, AgentID: "agent-cf-api0",
				AZ: "z1", Bootstrap: true, Deployment: "cf", IPs: []string{"10.0.1.30"},
				Job: "api", Index: 0, ID: "cf-api0-id", ProcessState: "running",
				State: "started", VMType: "medium", Ignore: false,
			},
			{
				VMCID: "vm-cf-uaa-0", Active: true, AgentID: "agent-cf-uaa0",
				AZ: "z1", Bootstrap: true, Deployment: "cf", IPs: []string{"10.0.1.40"},
				Job: "uaa", Index: 0, ID: "cf-uaa0-id", ProcessState: "running",
				State: "started", VMType: "medium", Ignore: false,
			},
			{
				VMCID: "vm-cf-doppler-0", Active: true, AgentID: "agent-cf-dop0",
				AZ: "z1", Bootstrap: true, Deployment: "cf", IPs: []string{"10.0.1.50"},
				Job: "doppler", Index: 0, ID: "cf-dop0-id", ProcessState: "running",
				State: "started", VMType: "small", Ignore: false,
			},
		},
		"redis": {
			{
				VMCID: "vm-redis-0", Active: true, AgentID: "agent-redis-0",
				AZ: "z1", Bootstrap: true, Deployment: "redis", IPs: []string{"10.0.4.10"},
				Job: "redis", Index: 0, ID: "redis-0-id", ProcessState: "running",
				State: "started", VMType: "medium", Ignore: false,
			},
			{
				VMCID: "vm-redis-1", Active: true, AgentID: "agent-redis-1",
				AZ: "z2", Bootstrap: false, Deployment: "redis", IPs: []string{"10.0.4.11"},
				Job: "redis", Index: 1, ID: "redis-1-id", ProcessState: "running",
				State: "started", VMType: "medium", Ignore: false,
			},
		},
		"mysql": {
			{
				VMCID: "vm-mysql-0", Active: true, AgentID: "agent-mysql-0",
				AZ: "z1", Bootstrap: true, Deployment: "mysql", IPs: []string{"10.0.5.10"},
				Job: "mysql", Index: 0, ID: "mysql-0-id", ProcessState: "running",
				State: "started", VMType: "large", Ignore: false,
			},
			{
				VMCID: "vm-mysql-1", Active: true, AgentID: "agent-mysql-1",
				AZ: "z2", Bootstrap: false, Deployment: "mysql", IPs: []string{"10.0.5.11"},
				Job: "mysql", Index: 1, ID: "mysql-1-id", ProcessState: "running",
				State: "started", VMType: "large", Ignore: false,
			},
			{
				VMCID: "vm-mysql-2", Active: true, AgentID: "agent-mysql-2",
				AZ: "z3", Bootstrap: false, Deployment: "mysql", IPs: []string{"10.0.5.12"},
				Job: "mysql", Index: 2, ID: "mysql-2-id", ProcessState: "running",
				State: "started", VMType: "large", Ignore: false,
			},
		},
	}
}

func defaultInstances() map[string][]Instance {
	return map[string][]Instance{
		"cf": {
			{
				AgentID: "agent-cf-dc0", AZ: "z1", Bootstrap: false, Deployment: "cf",
				Disk: "disk-cf-dc0", Expects: true, ID: "cf-dc0-id", IPs: []string{"10.0.1.10"},
				Job: "diego_cell", Index: 0, State: "running", VMType: "large", VMCID: "vm-cf-diego-cell-0",
				Processes: []Process{
					{Name: "rep", State: "running", Uptime: &Uptime{Seconds: 86400}, Memory: &ResourceUsage{Percent: 45.2, KB: 1024000}, CPU: &CPUUsage{Total: 12.5}},
					{Name: "garden", State: "running", Uptime: &Uptime{Seconds: 86400}, Memory: &ResourceUsage{Percent: 30.1, KB: 512000}, CPU: &CPUUsage{Total: 8.2}},
					{Name: "route_emitter", State: "running", Uptime: &Uptime{Seconds: 86400}, Memory: &ResourceUsage{Percent: 5.0, KB: 102400}, CPU: &CPUUsage{Total: 1.0}},
				},
			},
			{
				AgentID: "agent-cf-dc1", AZ: "z2", Bootstrap: false, Deployment: "cf",
				Disk: "disk-cf-dc1", Expects: true, ID: "cf-dc1-id", IPs: []string{"10.0.2.10"},
				Job: "diego_cell", Index: 1, State: "running", VMType: "large", VMCID: "vm-cf-diego-cell-1",
				Processes: []Process{
					{Name: "rep", State: "running", Uptime: &Uptime{Seconds: 86400}, Memory: &ResourceUsage{Percent: 42.0, KB: 980000}, CPU: &CPUUsage{Total: 10.0}},
					{Name: "garden", State: "running", Uptime: &Uptime{Seconds: 86400}, Memory: &ResourceUsage{Percent: 28.5, KB: 490000}, CPU: &CPUUsage{Total: 7.5}},
					{Name: "route_emitter", State: "running", Uptime: &Uptime{Seconds: 86400}, Memory: &ResourceUsage{Percent: 4.5, KB: 95000}, CPU: &CPUUsage{Total: 0.8}},
				},
			},
			{
				AgentID: "agent-cf-r0", AZ: "z1", Bootstrap: true, Deployment: "cf",
				Disk: "disk-cf-r0", Expects: true, ID: "cf-r0-id", IPs: []string{"10.0.1.20"},
				Job: "router", Index: 0, State: "running", VMType: "medium", VMCID: "vm-cf-router-0",
				Processes: []Process{
					{Name: "gorouter", State: "running", Uptime: &Uptime{Seconds: 172800}, Memory: &ResourceUsage{Percent: 20.0, KB: 256000}, CPU: &CPUUsage{Total: 15.0}},
					{Name: "route_registrar", State: "running", Uptime: &Uptime{Seconds: 172800}, Memory: &ResourceUsage{Percent: 2.0, KB: 25600}, CPU: &CPUUsage{Total: 0.5}},
				},
			},
			{
				AgentID: "agent-cf-api0", AZ: "z1", Bootstrap: true, Deployment: "cf",
				Disk: "disk-cf-api0", Expects: true, ID: "cf-api0-id", IPs: []string{"10.0.1.30"},
				Job: "api", Index: 0, State: "running", VMType: "medium", VMCID: "vm-cf-api-0",
				Processes: []Process{
					{Name: "cloud_controller_ng", State: "running", Uptime: &Uptime{Seconds: 259200}, Memory: &ResourceUsage{Percent: 35.0, KB: 450000}, CPU: &CPUUsage{Total: 8.0}},
					{Name: "nginx", State: "running", Uptime: &Uptime{Seconds: 259200}, Memory: &ResourceUsage{Percent: 5.0, KB: 64000}, CPU: &CPUUsage{Total: 2.0}},
				},
			},
		},
		"redis": {
			{
				AgentID: "agent-redis-0", AZ: "z1", Bootstrap: true, Deployment: "redis",
				Disk: "disk-redis-0", Expects: true, ID: "redis-0-id", IPs: []string{"10.0.4.10"},
				Job: "redis", Index: 0, State: "running", VMType: "medium", VMCID: "vm-redis-0",
				Processes: []Process{
					{Name: "redis-server", State: "running", Uptime: &Uptime{Seconds: 604800}, Memory: &ResourceUsage{Percent: 60.0, KB: 768000}, CPU: &CPUUsage{Total: 5.0}},
					{Name: "redis-sentinel", State: "running", Uptime: &Uptime{Seconds: 604800}, Memory: &ResourceUsage{Percent: 2.0, KB: 25600}, CPU: &CPUUsage{Total: 0.2}},
				},
			},
			{
				AgentID: "agent-redis-1", AZ: "z2", Bootstrap: false, Deployment: "redis",
				Disk: "disk-redis-1", Expects: true, ID: "redis-1-id", IPs: []string{"10.0.4.11"},
				Job: "redis", Index: 1, State: "running", VMType: "medium", VMCID: "vm-redis-1",
				Processes: []Process{
					{Name: "redis-server", State: "running", Uptime: &Uptime{Seconds: 604800}, Memory: &ResourceUsage{Percent: 55.0, KB: 704000}, CPU: &CPUUsage{Total: 4.5}},
					{Name: "redis-sentinel", State: "running", Uptime: &Uptime{Seconds: 604800}, Memory: &ResourceUsage{Percent: 1.8, KB: 23000}, CPU: &CPUUsage{Total: 0.2}},
				},
			},
		},
		"mysql": {
			{
				AgentID: "agent-mysql-0", AZ: "z1", Bootstrap: true, Deployment: "mysql",
				Disk: "disk-mysql-0", Expects: true, ID: "mysql-0-id", IPs: []string{"10.0.5.10"},
				Job: "mysql", Index: 0, State: "running", VMType: "large", VMCID: "vm-mysql-0",
				Processes: []Process{
					{Name: "pxc-mysql", State: "running", Uptime: &Uptime{Seconds: 1209600}, Memory: &ResourceUsage{Percent: 70.0, KB: 2048000}, CPU: &CPUUsage{Total: 20.0}},
					{Name: "galera-agent", State: "running", Uptime: &Uptime{Seconds: 1209600}, Memory: &ResourceUsage{Percent: 3.0, KB: 38400}, CPU: &CPUUsage{Total: 0.5}},
				},
			},
		},
	}
}

func defaultVariables() map[string][]Variable {
	return map[string][]Variable{
		"cf": {
			{ID: "var-1", Name: "cf_admin_password"},
			{ID: "var-2", Name: "uaa_admin_client_secret"},
			{ID: "var-3", Name: "router_ca"},
			{ID: "var-4", Name: "router_ssl"},
			{ID: "var-5", Name: "diego_instance_identity_ca"},
			{ID: "var-6", Name: "cc_db_encryption_key"},
		},
		"redis": {
			{ID: "var-10", Name: "redis_password"},
			{ID: "var-11", Name: "redis_tls_ca"},
		},
		"mysql": {
			{ID: "var-20", Name: "mysql_admin_password"},
			{ID: "var-21", Name: "pxc_galera_ca"},
			{ID: "var-22", Name: "mysql_server_certificate"},
		},
	}
}

func defaultTasks(now time.Time) map[int]*Task {
	return map[int]*Task{
		1: {
			ID: 1, State: "done", Description: "create deployment cf",
			Timestamp: now.Add(-24 * time.Hour).Unix(), Result: "Created", User: "admin", Deployment: "cf",
		},
		2: {
			ID: 2, State: "done", Description: "create deployment redis",
			Timestamp: now.Add(-20 * time.Hour).Unix(), Result: "Created", User: "admin", Deployment: "redis",
		},
		3: {
			ID: 3, State: "done", Description: "create deployment mysql",
			Timestamp: now.Add(-16 * time.Hour).Unix(), Result: "Created", User: "admin", Deployment: "mysql",
		},
		4: {
			ID: 4, State: "done", Description: "run errand smoke_tests",
			Timestamp: now.Add(-12 * time.Hour).Unix(), Result: "Errand completed successfully", User: "admin", Deployment: "cf",
		},
		5: {
			ID: 5, State: "error", Description: "run errand acceptance_tests",
			Timestamp: now.Add(-8 * time.Hour).Unix(), Result: "Error: Test failure in router tests", User: "admin", Deployment: "cf",
		},
		6: {
			ID: 6, State: "done", Description: "update deployment cf",
			Timestamp: now.Add(-4 * time.Hour).Unix(), Result: "Updated", User: "admin", Deployment: "cf",
		},
		7: {
			ID: 7, State: "done", Description: "snapshot deployment mysql",
			Timestamp: now.Add(-2 * time.Hour).Unix(), Result: "Snapshot created", User: "admin", Deployment: "mysql",
		},
		8: {
			ID: 8, State: "done", Description: "update cloud config",
			Timestamp: now.Add(-1 * time.Hour).Unix(), Result: "Updated", User: "admin",
		},
	}
}

func defaultStemcells() []Stemcell {
	return []Stemcell{
		{
			Name: "bosh-google-kvm-ubuntu-jammy-go_agent", OperatingSystem: "ubuntu-jammy",
			Version: "1.200", CID: "stemcell-uuid-1200",
			Deployments: []string{"cf", "redis", "mysql"},
		},
		{
			Name: "bosh-google-kvm-ubuntu-jammy-go_agent", OperatingSystem: "ubuntu-jammy",
			Version: "1.199", CID: "stemcell-uuid-1199",
			Deployments: []string{},
		},
		{
			Name: "bosh-google-kvm-ubuntu-bionic-go_agent", OperatingSystem: "ubuntu-bionic",
			Version: "1.150", CID: "stemcell-uuid-bionic-1150",
			Deployments: []string{},
		},
	}
}

func defaultReleases() []Release {
	return []Release{
		{Name: "cf-deployment", Version: "40.0.0", CommitHash: "abc123def", UncommittedChanges: false},
		{Name: "cf-deployment", Version: "39.0.0", CommitHash: "xyz789ghi", UncommittedChanges: false},
		{Name: "cflinuxfs4", Version: "1.50.0", CommitHash: "fs4abc123", UncommittedChanges: false},
		{Name: "diego", Version: "2.80.0", CommitHash: "diego80abc", UncommittedChanges: false},
		{Name: "garden-runc", Version: "1.28.0", CommitHash: "garden28xyz", UncommittedChanges: false},
		{Name: "redis", Version: "16.0.0", CommitHash: "redis16abc", UncommittedChanges: false},
		{Name: "pxc", Version: "0.42.0", CommitHash: "pxc42def", UncommittedChanges: false},
		{Name: "bpm", Version: "1.2.0", CommitHash: "bpm12ghi", UncommittedChanges: false},
		{Name: "os-conf", Version: "22.0.0", CommitHash: "osconf22jkl", UncommittedChanges: false},
	}
}

func defaultCloudConfig(now time.Time) *CloudConfig {
	return &CloudConfig{
		Properties: cloudConfigYAML(),
		CreatedAt:  now.Add(-1 * time.Hour).Format(time.RFC3339),
	}
}

func defaultRuntimeConfigs(now time.Time) []RuntimeConfig {
	return []RuntimeConfig{
		{
			Name:       "default",
			Properties: runtimeConfigYAML("default"),
			CreatedAt:  now.Add(-24 * time.Hour).Format(time.RFC3339),
		},
		{
			Name:       "dns",
			Properties: runtimeConfigYAML("dns"),
			CreatedAt:  now.Add(-48 * time.Hour).Format(time.RFC3339),
		},
	}
}

func defaultCPIConfig(now time.Time) *CPIConfig {
	return &CPIConfig{
		Properties: cpiConfigYAML(),
		CreatedAt:  now.Add(-72 * time.Hour).Format(time.RFC3339),
	}
}

func cloudConfigYAML() string {
	return `azs:
- name: z1
  cloud_properties:
    zone: us-central1-a
- name: z2
  cloud_properties:
    zone: us-central1-b
- name: z3
  cloud_properties:
    zone: us-central1-c

vm_types:
- name: small
  cloud_properties:
    machine_type: n1-standard-1
    root_disk_size_gb: 20
- name: medium
  cloud_properties:
    machine_type: n1-standard-2
    root_disk_size_gb: 50
- name: large
  cloud_properties:
    machine_type: n1-standard-4
    root_disk_size_gb: 100

networks:
- name: default
  type: manual
  subnets:
  - range: 10.0.0.0/16
    gateway: 10.0.0.1
    azs: [z1, z2, z3]
    dns: [8.8.8.8, 8.8.4.4]

compilation:
  workers: 5
  reuse_compilation_vms: true
  az: z1
  vm_type: medium
  network: default
`
}

func runtimeConfigYAML(name string) string {
	if name == "dns" {
		return `releases:
- name: bosh-dns
  version: 1.32.0

addons:
- name: bosh-dns
  jobs:
  - name: bosh-dns
    release: bosh-dns
`
	}
	return fmt.Sprintf(`releases:
- name: os-conf
  version: 22.0.0

addons:
- name: os-configuration
  jobs:
  - name: sysctl
    release: os-conf
    properties:
      sysctl:
      - net.ipv4.tcp_keepalive_time=120
`)
}

func cpiConfigYAML() string {
	return `cpis:
- name: gcp-cpi
  type: google
  properties:
    project: my-gcp-project
    default_zone: us-central1-a
`
}
