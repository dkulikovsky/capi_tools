package task

type Task struct {
	Owner      string
	ProjectId  string `yaml:"project_id"`
    Service    string
    Version    string
	Resources  Resources
	Volumes    map[string]Volume
	Command    string
    Ip         string
    Hostname   string
	StartHook  string `yaml:"start_hook"`
	StatusHook string `yaml:"status_hook"`
}

type Resources struct {
	Cpu  uint32
	Ram  uint64
	Net  int32 `yaml:",omitempty"`
	Disk int32
}

type Volume struct {
	Mount string
	Url   string
}


