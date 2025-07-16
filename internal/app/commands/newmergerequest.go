package commands

import (
	"fmt"
	"github.com/mkudelya/projectmaker/internal/app/types"
	"github.com/spf13/viper"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type NewMergeRequest struct {
	gitLabClient *gitlab.Client
}

func NewMergeRequestCommand(GitLabClient *gitlab.Client) *NewMergeRequest {
	return &NewMergeRequest{
		gitLabClient: GitLabClient,
	}
}

func (t *NewMergeRequest) Execute(projectID string, settings types.Settings, config *viper.Viper) error {
	if err := t.Validate(settings, config); err != nil {
		return err
	}

	projectPID := config.GetString("git.repository_id") + "/" + projectID

	reviewersIDs := make([]int, 0)
	authorsIDs := make([]int, 0)
	for _, user := range settings.GitUsers {
		if user.Type == types.GitReviewUserType {
			reviewersIDs = append(reviewersIDs, user.ID)
		} else {
			authorsIDs = append(authorsIDs, user.ID)
		}
	}

	// Create a new Merge Request
	m, _, err := t.gitLabClient.MergeRequests.CreateMergeRequest(projectPID, &gitlab.CreateMergeRequestOptions{
		SourceBranch:       gitlab.Ptr(settings.Branch),
		TargetBranch:       gitlab.Ptr(config.GetString("git.main_branch")),
		Title:              gitlab.Ptr(settings.TaskTitle),
		Description:        gitlab.Ptr(""),
		RemoveSourceBranch: gitlab.Ptr(true),
		ReviewerIDs:        &reviewersIDs,
		AssigneeIDs:        &authorsIDs,
	})

	if err == nil && m != nil {
		url := fmt.Sprintf("%s/%s/-/merge_requests/%d", config.GetString("git.server"), projectPID, m.IID)
		fmt.Printf("Pull request URL: %s\n", url)
	}

	return err
}

func (t *NewMergeRequest) Validate(settings types.Settings, config *viper.Viper) error {
	if settings.Branch == "" {
		return types.ErrEmptyProjectBranch
	}

	if settings.TaskTitle == "" {
		return types.ErrEmptyProjectTitle
	}

	if config.GetString("git.repository_id") == "" {
		return types.ErrEmptyGitRepositoryID
	}

	if !config.GetBool("git.enable") {
		return types.ErrGitDisabled
	}

	return nil
}
