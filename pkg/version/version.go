package version

import "fmt"

const (
	VERSION = "0.0.2"
)

func ShowVersion() (version string) {
	fmt.Println(VERSION)
	return
}
