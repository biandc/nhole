package tools

import (
	"github.com/biandc/nhole/pkg/log"
	"github.com/biandc/nhole/pkg/version"
)

func PrintLogo() {
	logo := `███▄▄▄▄      ▄█    █▄     ▄██████▄   ▄█          ▄████████ 
███▀▀▀██▄   ███    ███   ███    ███ ███         ███    ███ 
███   ███   ███    ███   ███    ███ ███         ███    █▀  
███   ███  ▄███▄▄▄▄███▄▄ ███    ███ ███        ▄███▄▄▄     
███   ███ ▀▀███▀▀▀▀███▀  ███    ███ ███       ▀▀███▀▀▀     
███   ███   ███    ███   ███    ███ ███         ███    █▄  
███   ███   ███    ███   ███    ███ ███▌    ▄   ███    ███ 
 ▀█   █▀    ███    █▀     ▀██████▀  █████▄▄██   ██████████ 
                                    ▀`
	log.Info("VERSION: %s\n%s", version.VERSION, logo)
}
