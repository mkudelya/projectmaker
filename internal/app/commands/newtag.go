package commands

import (
	"fmt"
	"github.com/mkudelya/projectmaker/internal/app/types"
	"github.com/mkudelya/projectmaker/internal/app/utils"
	"github.com/rotisserie/eris"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type NewTag struct {
}

func NewTagCommand() *NewTag {
	return &NewTag{}
}

func (t *NewTag) Execute(projectId string, settings types.Settings, config *viper.Viper) error {
	if err := t.Validate(settings, config); err != nil {
		return err
	}

	os.Chdir(utils.PathByProjectID(settings, projectId))

	fmt.Printf("Project '%s' checkout to source branch\n", projectId)
	cmd := exec.Command("git", "checkout", config.GetString("git.main_branch"))
	if _, err := utils.ProcessExecResult(cmd); err != nil {
		return eris.Wrapf(err, "failed to checkout branch '%s'", config.GetString("git.main_branch"))
	}

	fmt.Printf("Project '%s' try to remore update\n", projectId)
	cmd = exec.Command("git", "remote", "update")
	_, err := utils.ProcessExecResult(cmd)
	if err != nil {
		return eris.Wrapf(err, "failed to remote update in branch '%s'", config.GetString("git.main_branch"))
	}

	fmt.Printf("Project '%s' try to pull\n", projectId)
	cmd = exec.Command("git", "pull")
	_, err = utils.ProcessExecResult(cmd)
	if err != nil {
		return eris.Wrapf(err, "failed to pull in branch '%s'", config.GetString("git.main_branch"))
	}

	fmt.Printf("Project '%s' try to get latest tag\n", projectId)
	cmd = exec.Command("git", "tag", "-l", "--sort=v:refname")

	output, err := utils.ProcessExecResult(cmd)
	if err != nil {
		return eris.Wrapf(err, "failed get latest tag in branch '%s'", config.GetString("git.main_branch"))
	}

	output = strings.Trim(output, "\r\n")

	re := regexp.MustCompile(`(v)(\d+)\.(\d+)\.(\d+)(.*)`)
	match := re.FindAllStringSubmatch(output, -1)
	if len(match) == 0 {
		fmt.Println(output)
		return eris.Errorf("failed get latest tag in branch '%s'", config.GetString("git.main_branch"))
	}

	latestMatch := len(match) - 1

	fmt.Printf("Project '%s' latest tag %s\n", projectId, match[latestMatch][0])
	if len(match[0]) < 5 {
		return eris.Errorf("failed parse latest tag '%s' in branch '%s'", match[latestMatch][0], config.GetString("git.main_branch"))
	}
	minorVersion := match[latestMatch][3]
	patchVersion := match[latestMatch][4]

	patchVersionInt, err := strconv.Atoi(patchVersion)
	if err != nil {
		return eris.Wrapf(err, "failed convert patch version '%s' to int", patchVersion)
	}

	minorVersionInt, err := strconv.Atoi(minorVersion)
	if err != nil {
		return eris.Wrapf(err, "failed convert minor version '%s' to int", patchVersion)
	}

	patchVersionInt++

	if patchVersionInt > 10 {
		minorVersionInt++
		patchVersionInt = 0
	}

	newTag := fmt.Sprintf("%s%s.%d.%d%s", match[latestMatch][1], match[latestMatch][2], minorVersionInt, patchVersionInt, match[latestMatch][5])
	fmt.Printf("Project '%s' new tag %s\n", projectId, newTag)
	var answer string
	fmt.Print("Do you want to create new tag? (Y/n/[custom name]): ")
	fmt.Scanf("%s", &answer)

	if strings.ToLower(answer) != "y" && answer != "" {
		if strings.ToLower(answer) == "n" {
			return types.ErrIsNotAgreeWithTagCreation
		} else {
			newTag = answer
		}
	}
	
	fmt.Printf("Project '%s' try to create new tag %s\n", projectId, newTag)
	cmd = exec.Command("git", "tag", newTag)
	_, err = utils.ProcessExecResult(cmd)
	if err != nil {
		return eris.Wrapf(err, "failed to create new tag in branch '%s'", config.GetString("git.main_branch"))
	}

	fmt.Printf("Project '%s' try to push new tag %s\n", projectId, newTag)
	cmd = exec.Command("git", "push", config.GetString("git.remote_name"), newTag)
	_, err = utils.ProcessExecResult(cmd)
	if err != nil {
		return eris.Wrapf(err, "failed to push new tag in branch '%s'", config.GetString("git.main_branch"))
	}

	return nil
}

func (t *NewTag) Validate(settings types.Settings, config *viper.Viper) error {
	if config.GetString("git.repository_id") == "" {
		return types.ErrEmptyGitRepositoryID
	}

	return nil
}
