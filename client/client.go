package client

import (
	"context"

	"github.com/biandc/nhole/pkg/config"
	"github.com/biandc/nhole/pkg/control"
	"github.com/biandc/nhole/pkg/log"
	"github.com/biandc/nhole/pkg/tools"
)

const (
	LogWay          = "console"
	LogFile         = ""
	LogLevel        = "debug"
	LogMaxdays      = 0
	LogDisableColor = false
)

func Run(cfg *config.ClientCfg) (err error) {
	log.InitLog(LogWay, LogFile, LogLevel, LogMaxdays, LogDisableColor)

	var (
		clienter *control.ControlClient
	)
	ctx := context.WithValue(context.Background(), "cfg", cfg)
	clienter, err = control.NewControlClienter(ctx, cfg.Server.Ip, cfg.Server.ControlPort)
	if err != nil {
		return
	}
	log.Info("nhole-client start ...")
	clienter.Run()
	tools.ExitClear(clienter, "nhole-client exit ...")
	return
}
