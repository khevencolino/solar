package debug

import "fmt"

var Enabled bool = false

func Printf(format string, args ...interface{}) {
	if Enabled {
		fmt.Printf(format, args...)
	}
}

func Println(args ...interface{}) {
	if Enabled {
		fmt.Println(args...)
	}
}

func Print(args ...interface{}) {
	if Enabled {
		fmt.Print(args...)
	}
}
