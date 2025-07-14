package commands

import (
	"fmt"
	"github.com/mkudelya/projectmaker/internal/app/types"
	"github.com/mkudelya/projectmaker/internal/app/utils"
	"github.com/rotisserie/eris"
	"github.com/spf13/viper"
	"os"
	"os/exec"
)

type NewTask struct {
}

func NewTaskCommand() *NewTask {
	return &NewTask{}
}

func (t *NewTask) Execute(projectAlias string, settings types.Settings, config *viper.Viper) error {
	if err := t.Validate(settings); err != nil {
		return err
	}

	os.Chdir(utils.PathByProjectAlias(settings, projectAlias))

	fmt.Printf("Project '%s' checkout to source branch\n", projectAlias)
	cmd := exec.Command("git", "checkout", config.GetString("git.source_branch"))
	if _, err := utils.ProcessExecResult(cmd); err != nil {
		return eris.Wrapf(err, "failed to checkout branch '%s'", config.GetString("git.source_branch"))
	}

	fmt.Printf("Project '%s' try to remore update\n", projectAlias)
	cmd = exec.Command("git", "remote", "update")
	_, err := utils.ProcessExecResult(cmd)
	if err != nil {
		return eris.Wrapf(err, "failed to remote update in branch '%s'", config.GetString("git.source_branch"))
	}

	fmt.Printf("Project '%s' create new branch\n", projectAlias)
	cmd = exec.Command("git", "checkout", "-b", settings.Branch)
	if _, err := utils.ProcessExecResult(cmd); err != nil {
		return eris.Wrapf(err, "failed to create new branch '%s'", settings.Branch)
	}

	fmt.Printf("Project '%s' push new branch to remote\n", projectAlias)
	cmd = exec.Command("git", "push", config.GetString("git.remote_name"), settings.Branch)
	if _, err := utils.ProcessExecResult(cmd); err != nil {
		return eris.Wrapf(err, "failed to push new branch '%s' to remote", settings.Branch)
	}

	return nil
}

func (t *NewTask) Validate(settings types.Settings) error {
	if settings.Branch == "" {
		return types.ErrEmptyProjectBranch
	}

	return nil
}
