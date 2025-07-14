package utils

import (
	"fmt"
	"github.com/mkudelya/projectmaker/internal/app/types"
	"os"
	"os/exec"
)

func PathByProjectAlias(settings types.Settings, alias string) string {
	for _, project := range settings.ConfigProjects {
		if project.Alias == alias {
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

func IsExistGitUsersFile() bool {
	_, err := os.Stat(types.GitReviewUsersFileName)
	return !os.IsNotExist(err)
}

func IsAllReviewersFound(reviewersFileList []types.GitUser, reviewersConfigList []string) bool {
	if len(reviewersFileList) == 0 {
		return false
	}

	for _, reviewerName := range reviewersConfigList {
		var isFound bool
		for _, reviewUser := range reviewersFileList {
			if reviewerName == reviewUser.UserName {
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
