package server

import (
	"context"

	"github.com/biandc/nhole/pkg/config"
	"github.com/biandc/nhole/pkg/control"
	"github.com/biandc/nhole/pkg/log"
	"github.com/biandc/nhole/pkg/tools"
)

var (
	LogWay          string
	LogFile         string
	LogLevel        string
	LogDisableColor bool
)

func Run(cfg *config.ServerCfg) (err error) {
	tools.PrintLogo()
	log.InitLog(LogWay, LogFile, LogLevel, LogDisableColor)

	var (
		server *control.ControlServ
	)
	ctx := context.WithValue(context.Background(), "cfg", cfg)
	server, err = control.NewControlServer(ctx, cfg.Server.Ip, cfg.Server.ControlPort)
	if err != nil {
		return
	}
	log.Info("nhole-server start ...")
	server.Run()
	tools.ExitClear(server, "nhole-server exit ...")
	return
}
