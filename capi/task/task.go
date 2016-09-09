package task

type Task struct {
	Owner     string
	ProjectId string `yaml:"project_id"`
    Service   string
    Version   string
	Spec      Spec
}

type Spec struct {
	Resources  Resources
	Volumes    map[string]Volume
	Command    string
    Ip         string
    Hostname   string
	StartHook  string `yaml:"start_hook"`
	StatusHook string `yaml:"status_hook"`
}

type Resources struct {
	Cpu  int32
	Ram  int32
	Net  int32 `yaml:",omitempty"`
	Disk int32
}

type Volume struct {
	Mount string
	Url   string
}


