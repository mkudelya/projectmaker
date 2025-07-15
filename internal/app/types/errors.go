package types

import "errors"

var ErrInvalidCommand = errors.New("invalid command in command line")
var ErrEmptyProjects = errors.New("empty projects in command line")
var ErrEmptyProjectBranch = errors.New("empty project branch in command line")
var ErrEmptyProjectTitle = errors.New("empty project title in command line")
var ErrIsNotEnoughArgv = errors.New("you must specify command and project names in command line")
var ErrIsNotExistProjectID = errors.New("project id is not exist in command line")

var ErrEmptyProjectID = errors.New("empty project ID in config file")
var ErrEmptyGitSourceBranch = errors.New("empty git source branch in config file")

var ErrIsNotExistProjectPath = errors.New("project path is not exist")
var ErrIsNotProjectDirectory = errors.New("project path is not directory")

var ErrIsNotAgreeWithTagCreation = errors.New("tag creation rejected")
var ErrEmptyGitRepositoryID = errors.New("empty git repository ID in config file")
var ErrEmptyGitRemoteName = errors.New("empty git remote name in config file")
var ErrEmptyGitAuthors = errors.New("empty git authors users")
var ErrEmptyGitReviewers = errors.New("empty git reviewers users")
var ErrGitDisabled = errors.New("git is disabled in config file")
