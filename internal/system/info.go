package system

import (
	"github.com/kodelint/shell-agent/internal/logger"
	"github.com/spf13/viper"
	"runtime"
)

var log = logger.GetLogger()

type SystemInfo struct{}

type Info struct {
	OS         string
	Arch       string
	GoVersion  string
	ConfigFile string
	Debug      bool
	Verbose    bool
}

func NewSystemInfo() *SystemInfo {
	log.Debugf("Creating new system info instance")
	return &SystemInfo{}
}

func (s *SystemInfo) GetInfo() *Info {
	log.Debugf("Getting system info for new instance")
	return &Info{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		GoVersion:  runtime.Version(),
		ConfigFile: viper.ConfigFileUsed(),
		Debug:      viper.GetBool("debug"),
		Verbose:    viper.GetBool("verbose"),
	}
}
