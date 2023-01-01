package config

import (
	"os"

	"github.com/biandc/nhole/pkg/tools"
	"gopkg.in/yaml.v3"
)

type Service struct {
	Ip          string `yaml:"ip"`
	Port        int    `yaml:"port"`
	ForwardPort int    `yaml:"forward_port"`
}

type ClientCfg struct {
	Server   Server     `yaml:"server"`
	Services []*Service `yaml:"service"`
}

func (c *ClientCfg) Validate() (err error) {
	err = tools.ValidateIp(c.Server.Ip)
	if err != nil {
		return
	}
	err = tools.ValidatePort(c.Server.ControlPort)
	if err != nil {
		return
	}
	for _, value := range c.Services {
		err = tools.ValidateIp(value.Ip)
		if err != nil {
			return
		}
		err = tools.ValidatePort(value.Port)
		if err != nil {
			return
		}
		err = tools.ValidatePort(value.ForwardPort)
		if err != nil {
			return
		}
	}
	return
}

func UnmarshalClientCfg(content []byte) (cfg *ClientCfg, err error) {
	cfg = &ClientCfg{}
	err = yaml.Unmarshal(content, cfg)
	if err != nil {
		return
	}
	err = cfg.Validate()
	return
}

func UnmarshalClientCfgByFile(file string) (cfg *ClientCfg, err error) {
	var content []byte
	content, err = os.ReadFile(file)
	if err != nil {
		return
	}
	cfg, err = UnmarshalClientCfg(content)
	return
}
