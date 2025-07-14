package types

type Project struct {
	Alias string `mapstructure:"alias"`
	Name  string `mapstructure:"name"`
	Path  string `mapstructure:"path"`
}

type Settings struct {
	ConfigProjects map[string]Project `mapstructure:"projects"`

	Command             string
	UserProjectsAliases []string
	Branch              string
	TaskTitle           string
	GitUsers            []GitUser
}

type GitUser struct {
	ID         int    `json:"id"`
	UserName   string `json:"username"`
	ServerName string `json:"serverName"`
}
