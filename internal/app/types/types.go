package types

type Project struct {
	ProjectID string `mapstructure:"project_id"`
	Path      string `mapstructure:"path"`
}

type Settings struct {
	ConfigProjects map[string]Project `mapstructure:"projects"`

	Command        string
	UserProjectIDs []string
	Branch         string
	TaskTitle      string
	GitUsers       []GitUser
}

type GitUser struct {
	ID         int    `json:"id"`
	UserName   string `json:"username"`
	ServerName string `json:"serverName"`
	Type       string `json:"type"` //author or reviewer
}
