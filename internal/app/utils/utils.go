package utils

import (
	"fmt"
	"github.com/mkudelya/projectmaker/internal/app/types"
	"github.com/spf13/viper"
	"os/exec"
	"slices"
)

func PathByProjectID(settings types.Settings, ID string) string {
	for _, project := range settings.ConfigProjects {
		if project.ProjectID == ID {
			return project.Path
		}
	}

	return ""
}

func ProcessExecResult(cmd *exec.Cmd) (string, error) {
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		return string(output), err
	}

	return string(output), nil
}

func IsAllUsersFound(gitUsers []types.GitUser, gitConfigUsers []string) bool {
	if len(gitUsers) == 0 {
		return false
	}

	for _, reviewerName := range gitConfigUsers {
		var isFound bool
		for _, user := range gitUsers {
			if reviewerName == user.UserName {
				isFound = true
				break
			}
		}
		if !isFound {
			return false
		}
	}

	return true
}

func SetTypeGitUser(config *viper.Viper, gitUsers []types.GitUser) []types.GitUser {
	authors := config.GetStringSlice("pull_request.authors")
	reviwers := config.GetStringSlice("pull_request.reviewers")

	for i := range gitUsers {
		if slices.Contains(authors, gitUsers[i].UserName) {
			gitUsers[i].Type = types.GitAuthorUserType
			continue
		}

		if slices.Contains(reviwers, gitUsers[i].UserName) {
			gitUsers[i].Type = types.GitReviewUserType
			continue
		}
	}

	return gitUsers
}
