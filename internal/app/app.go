package app

import (
	"fmt"
	"github.com/spf13/viper"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type App struct {
	GitLabClient *gitlab.Client
	Config       *viper.Viper
}

func (a *App) NewApp() *App {
	a.initApp()

	return a
}

func (a *App) initApp() {
	a.initConfig()
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
	a.GitLabClient, err = gitlab.NewClient("glpat-GxARLo8JwsWcyypVME--", gitlab.WithBaseURL("http://gitlab.svc.aku.com"))

	if err != nil {

	}

	username := "HannenkoAO"
	users, _, err := a.GitLabClient.Users.ListUsers(&gitlab.ListUsersOptions{Search: &username})

	//fmt.Println(err)
	//
	//fmt.Println(response)

	fmt.Println("users count: ", len(users))

	for _, user := range users {
		fmt.Println(user.ID, user.Username)
	}
}
