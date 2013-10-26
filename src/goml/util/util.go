package util

import "fmt"

func Assert(condition bool, v ...interface{}) {
	if !condition {
		panic(fmt.Sprint(v...))
	}
}
