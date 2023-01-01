package config

import (
	"os"

	"github.com/biandc/nhole/pkg/tools"
	"gopkg.in/yaml.v3"
)

type Server struct {
	Ip          string `yaml:"ip"`
	ControlPort int    `yaml:"control_port"`
}

type ServerCfg struct {
	Server Server `yaml:"server"`
}

func (s *ServerCfg) Validate() (err error) {
	err = tools.ValidateIp(s.Server.Ip)
	if err != nil {
		return
	}
	err = tools.ValidatePort(s.Server.ControlPort)
	return
}

func UnmarshalServerCfg(content []byte) (cfg *ServerCfg, err error) {
	cfg = &ServerCfg{}
	err = yaml.Unmarshal(content, cfg)
	if err != nil {
		return
	}
	err = cfg.Validate()
	return
}

func UnmarshalServerCfgByFile(file string) (cfg *ServerCfg, err error) {
	var content []byte
	content, err = os.ReadFile(file)
	if err != nil {
		return
	}
	cfg, err = UnmarshalServerCfg(content)
	return
}
