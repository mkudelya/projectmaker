package types

import "errors"

var ErrInvalidCommand = errors.New("invalid command in command line")
var ErrEmptyProjects = errors.New("empty projects in command line")
var ErrEmptyProjectBranch = errors.New("empty project branch in command line")
var ErrIsNotEnoughArgv = errors.New("you must specify command and project names in command line")
var ErrIsNotExistProjectAlias = errors.New("project alias is not exist in command line")

var ErrEmptyProjectAlias = errors.New("empty project alias in config file")
var ErrEmptyGitSourceBranch = errors.New("empty git source branch in config file")

var ErrIsNotExistProjectPath = errors.New("project path is not exist")
var ErrIsNotProjectDirectory = errors.New("project path is not directory")

var ErrIsNotAgreeWithTagCreation = errors.New("tag creation rejected")
