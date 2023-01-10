package version

import "fmt"

const (
	VERSION = "0.0.8"
)

func ShowVersion() (version string) {
	fmt.Println(VERSION)
	return
}
