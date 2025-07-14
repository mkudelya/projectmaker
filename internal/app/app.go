package app

import (
	"encoding/json"
	"fmt"
	"github.com/mkudelya/projectmaker/internal/app/commands"
	"github.com/mkudelya/projectmaker/internal/app/types"
	"github.com/mkudelya/projectmaker/internal/app/utils"
	"github.com/rotisserie/eris"
	"github.com/spf13/viper"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"os"
	"slices"
	"strings"
)

type App struct {
	GitLabClient *gitlab.Client
	Config       *viper.Viper
	Settings     types.Settings
}

func NewApp() *App {
	return &App{}
}

func (a *App) InitApp(argv []string) error {
	a.initConfig()
	a.initGitLabClient()

	err := a.initGitReviewUsers()
	if err != nil {
		return eris.Wrap(err, "failed to init git review users")
	}

	err = a.initArgv(argv)
	if err == nil {
		return eris.Wrap(err, "failed to init argv")
	}

	return nil
}

func (a *App) initConfig() {
	a.Config = viper.New()
	a.Config.SetConfigName("config")
	a.Config.SetConfigType("yaml")
	a.Config.AddConfigPath(".")
	err := a.Config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func (a *App) initGitLabClient() {
	var err error

	url, token := a.Config.GetString("git.server"), a.Config.GetString("git.token")
	if url == "" || token == "" {
		return
	}

	a.GitLabClient, err = gitlab.NewClient(token, gitlab.WithBaseURL(url))
	if err != nil {
		panic(fmt.Errorf("gitlab error: %w", err))
	}
}

func (a *App) initGitReviewUsers() error {
	reviewerConfigList := a.Config.GetStringSlice("pull_request_reviewers")
	var err error
	if len(reviewerConfigList) == 0 {
		return nil
	}

	var reviewerList []types.GitUser
	reviewerList = a.GitReviewUsersFromFile()

	isFound := utils.IsAllReviewersFound(reviewerList, reviewerConfigList)

	if !isFound {
		reviewerList, err = a.GitReviewersUsers(reviewerConfigList)
		if err != nil {
			return eris.Wrapf(err, "failed to get git reviewers users from git server")
		}

		err = a.CreateGitReviewUsersFile(reviewerList)
		if err != nil {
			return eris.Wrapf(err, "failed to create git review users file")
		}
	}

	a.Settings.GitUsers = reviewerList

	return nil
}

func (a *App) GitReviewersUsers(reviewerConfigList []string) ([]types.GitUser, error) {
	gitUsers := make([]types.GitUser, 0)

	for _, userName := range reviewerConfigList {
		trimmedName := strings.TrimSpace(userName)
		users, _, err := a.GitLabClient.Users.ListUsers(&gitlab.ListUsersOptions{Search: &trimmedName})

		if err != nil {
			return gitUsers, eris.Wrapf(err, "failed to get git users by userName '%s'", trimmedName)
		}

		var isFound bool
		for _, user := range users {
			if user.Username == userName {
				gitUsers = append(gitUsers, types.GitUser{
					ID:         user.ID,
					UserName:   user.Username,
					ServerName: a.Config.GetString("git.server"),
				})
				isFound = true
			}
		}

		if !isFound {
			return gitUsers, eris.Errorf("User '%s' not found in gitlab server\n", userName)
		}
	}

	return gitUsers, nil
}

func (a *App) CreateGitReviewUsersFile(reviewerList []types.GitUser) error {
	gitUsersJson, err := json.Marshal(reviewerList)
	if err != nil {
		return eris.Wrapf(err, "failed to marshal git users to json")
	}

	file, err := os.Create(types.GitReviewUsersFileName)
	if err != nil {
		return eris.Wrapf(err, "failed to create git review users file '%s'", types.GitReviewUsersFileName)
	}

	_, err = file.Write(gitUsersJson)
	if err != nil {
		return eris.Wrapf(err, "failed to write git review users to file '%s'", types.GitReviewUsersFileName)
	}

	return nil
}

func (a *App) GitReviewUsersFromFile() []types.GitUser {
	gitUsers := make([]types.GitUser, 0)

	data, err := os.ReadFile(types.GitReviewUsersFileName)
	if err != nil {
		return gitUsers
	}

	if len(data) == 0 {
		return gitUsers
	}

	json.Unmarshal(data, &gitUsers)

	return gitUsers
}

func (a *App) initArgv(argv []string) error {
	if len(argv) >= 4 {
		a.Settings.Branch = argv[3]
	}

	if len(argv) >= 5 {
		a.Settings.TaskTitle = argv[4]
	}

	if len(argv) < 3 {
		return types.ErrIsNotEnoughArgv
	}

	if a.Config.GetString("git.source_branch") == "" {
		return types.ErrEmptyGitSourceBranch
	}

	command := argv[1]
	if slices.Contains(types.AllowedCommands, command) {
		a.Settings.Command = command
	} else {
		return eris.Wrapf(types.ErrInvalidCommand, "%s", command)
	}

	a.Settings.UserProjectsAliases = strings.Split(argv[2], ",")
	if len(a.Settings.UserProjectsAliases) == 0 {
		return types.ErrEmptyProjects
	}

	if err := a.Config.Unmarshal(&a.Settings); err != nil {
		panic(fmt.Errorf("cannot parse 'projects' config section: %w", err))
	}

	for i, project := range a.Settings.ConfigProjects {
		if project.Alias == "" {
			return eris.Wrapf(types.ErrEmptyProjectAlias, "'%s'", project.Name)
		}

		info, err := os.Stat(project.Path)
		if os.IsNotExist(err) {
			return eris.Wrapf(types.ErrIsNotExistProjectPath, "'%s'", project.Path)
		}

		if !info.IsDir() {
			return eris.Wrapf(types.ErrIsNotProjectDirectory, "'%s'", project.Path)
		}

		path := strings.TrimSuffix(project.Path, string(os.PathSeparator))
		p := a.Settings.ConfigProjects[i]
		p.Path = path + string(os.PathSeparator)
		a.Settings.ConfigProjects[i] = p
	}

	for _, userAlias := range a.Settings.UserProjectsAliases {
		isExist := false
		for _, project := range a.Settings.ConfigProjects {
			if project.Alias == userAlias {
				isExist = true
			}
		}
		if !isExist {
			return eris.Wrapf(types.ErrIsNotExistProjectAlias, "%s", userAlias)
		}
	}

	return nil
}

func (a *App) ExecuteCommand() error {
	var command commands.Command

	switch a.Settings.Command {
	case "newtask":
		command = commands.NewTaskCommand()
	case "newtag":
		command = commands.NewTagCommand()
	case "newmergerequest":
		command = commands.NewMergeRequestCommand()
	default:
		return eris.Wrapf(types.ErrInvalidCommand, "%s", a.Settings.Command)
	}

	for _, alias := range a.Settings.UserProjectsAliases {
		fmt.Printf("Project '%s' start command execute\n", alias)

		err := command.Execute(alias, a.Settings, a.Config)
		if err != nil {
			return eris.Wrapf(err, "failed to execute command in project '%s'", alias)
		}

		fmt.Printf("Project '%s' command executed successfully\n", alias)
	}

	return nil
}
