package practice

import (
	"fmt"
)

func Demo() {
	var fs []func()

	for i := 0; i < 3; i++ {
		fs = append(fs, func(v int) func() {
			return func() { fmt.Println(v) }
		}(i))
	}

	for _, f := range fs {
		f()
	}
}
