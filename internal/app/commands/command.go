package commands

import (
	"github.com/mkudelya/projectmaker/internal/app/types"
	"github.com/spf13/viper"
)

type Command interface {
	Execute(projectID string, settings types.Settings, config *viper.Viper) error
	Validate(settings types.Settings, config *viper.Viper) error
}
