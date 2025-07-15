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

	if a.Config.GetBool("git.enable") {
		a.initGitLabClient()

		err := a.initGitUsers()
		if err != nil {
			return eris.Wrap(err, "failed to init git review users")
		}
	}

	if len(argv) == 1 {
		a.Usage()
		return nil
	}

	err := a.initArgv(argv)
	if err != nil {
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

	a.GitLabClient, err = gitlab.NewClient(
		a.Config.GetString("git.token"),
		gitlab.WithBaseURL(a.Config.GetString("git.server")),
	)

	if err != nil {
		panic(fmt.Errorf("gitlab error: %w", err))
	}
}

func (a *App) initGitUsers() error {
	gitConfigUsers := a.Config.GetStringSlice("pull_request.authors")
	gitConfigUsers = append(gitConfigUsers, a.Config.GetStringSlice("pull_request.reviewers")...)
	var err error
	if len(gitConfigUsers) == 0 {
		return nil
	}

	var gitUsers []types.GitUser
	gitUsers = a.GitUsersFromFile()

	isFound := utils.IsAllUsersFound(gitUsers, gitConfigUsers)

	if !isFound {
		gitUsers, err = a.GitReviewersUsers(gitConfigUsers)
		if err != nil {
			return eris.Wrapf(err, "failed to get git users from git server")
		}

		err = a.CreateGitUsersFile(gitUsers)
		if err != nil {
			return eris.Wrapf(err, "failed to create git users file")
		}
	}

	a.Settings.GitUsers = utils.SetTypeGitUser(a.Config, gitUsers)

	var authorsCount, reviewersCount int
	for _, user := range a.Settings.GitUsers {
		if user.Type == types.GitAuthorUserType {
			authorsCount++
		}

		if user.Type == types.GitReviewUserType {
			reviewersCount++
		}
	}

	if authorsCount == 0 {
		return types.ErrEmptyGitAuthors
	}

	if reviewersCount == 0 {
		return types.ErrEmptyGitReviewers

	}

	return nil
}

func (a *App) GitReviewersUsers(gitConfigUsers []string) ([]types.GitUser, error) {
	gitUsers := make([]types.GitUser, 0)

	if a.GitLabClient == nil {
		return gitUsers, eris.New("gitlab client is not initialized")
	}

	for _, userName := range gitConfigUsers {
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

func (a *App) CreateGitUsersFile(reviewerList []types.GitUser) error {
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

func (a *App) GitUsersFromFile() []types.GitUser {
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

	if a.Config.GetString("git.main_branch") == "" {
		return types.ErrEmptyGitSourceBranch
	}

	command := argv[1]
	if slices.Contains(types.AllowedCommands, command) {
		a.Settings.Command = command
	} else {
		return eris.Wrapf(types.ErrInvalidCommand, "%s", command)
	}

	a.Settings.UserProjectIDs = strings.Split(argv[2], ",")
	if len(a.Settings.UserProjectIDs) == 0 {
		return types.ErrEmptyProjects
	}

	if err := a.Config.Unmarshal(&a.Settings); err != nil {
		panic(fmt.Errorf("cannot parse 'projects' config section: %w", err))
	}

	for i, project := range a.Settings.ConfigProjects {
		if project.ProjectID == "" {
			return eris.Wrapf(types.ErrEmptyProjectID, "'%s'", project.ProjectID)
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

	for _, projectID := range a.Settings.UserProjectIDs {
		isExist := false
		for _, project := range a.Settings.ConfigProjects {
			if project.ProjectID == projectID {
				isExist = true
			}
		}
		if !isExist {
			return eris.Wrapf(types.ErrIsNotExistProjectID, "%s", projectID)
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
		command = commands.NewMergeRequestCommand(a.GitLabClient)
	default:
		return eris.Wrapf(types.ErrInvalidCommand, "%s", a.Settings.Command)
	}

	for _, projectID := range a.Settings.UserProjectIDs {
		fmt.Printf("Project '%s' start command execute\n", projectID)

		err := command.Execute(projectID, a.Settings, a.Config)
		if err != nil {
			return eris.Wrapf(err, "failed to execute command in project '%s'", projectID)
		}

		fmt.Printf("Project '%s' command executed successfully\n", projectID)
	}

	return nil
}

func (a *App) Usage() {
	fmt.Printf("Usage: COMMAND PROJECT_IDS [PROJECT_BRANCH] \"[PROJECT_MR_TITLE]\"\n" +
		"COMMAND:\n" +
		"  newtask            Create new task branch from main branch\n" +
		"  newtag             Create new tag and push it\n" +
		"  newmergerequest    Create new merge request from [PROJECT BRANCH]\n" +
		"PROJECT_IDS: list of project ids via comma\n" +
		"PROJECT_BRANCH: branch name for new task or merge request\n" +
		"PROJECT_MR_TITLE: title for new merge request\n",
	)
}
