package version

import "fmt"

const (
	VERSION = "0.0.6"
)

func ShowVersion() (version string) {
	fmt.Println(VERSION)
	return
}
