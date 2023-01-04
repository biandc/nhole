package version

import "fmt"

const (
	VERSION = "0.0.4"
)

func ShowVersion() (version string) {
	fmt.Println(VERSION)
	return
}
