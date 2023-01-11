package client

import (
	"context"
	"time"

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

func Run(cfg *config.ClientCfg) (err error) {
	tools.PrintLogo()
	log.InitLog(LogWay, LogFile, LogLevel, LogDisableColor)

	var (
		clienter *control.ControlClient
	)
	ctx := context.WithValue(context.Background(), "cfg", cfg)
	clienter, err = control.NewControlClienter(ctx, cfg.Server.Ip, cfg.Server.ControlPort)
	if err != nil {
		return
	}
	go func() {
		for {
			err := clienter.Init()
			if err != nil {
				log.Error(err.Error())
				time.Sleep(1 * time.Second)
				continue
			}
			clienter.Run()
		}
	}()
	log.Info("nhole-client start ...")
	tools.ExitClear(clienter, "nhole-client exit ...")
	return
}
