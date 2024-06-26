package context

import (
	"azdo-dash/config"
)

type ProgramContext struct {
	Config     *config.Config
	ConfigPath string
}
