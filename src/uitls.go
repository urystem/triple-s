package src

import "os"

func Error(s string) {
	os.Stdout.WriteString(s)
}
