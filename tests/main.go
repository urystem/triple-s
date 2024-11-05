package main

import (
	"fmt"
	"regexp"
)

func main() {
	s := "dsd*"
	fmt.Println(regexp.MustCompile(`^[-.a-z\d]{3,63}$`).MatchString(s))
}

// func check(b string) (p bool) { // check bucket name
// 	if len(b) < 3 || len(b) > 63 || b[0] == '.' || b[len(b)-1] == '.' || b[0] == '-' || b[len(b)-1] == '-' {
// 		return false
// 	}
// 	for i := range b {
// 		if b[i] >= 'a' && b[i] <= 'z' {
// 			p = true
// 		} else if (b[i] < '0' || b[i] > '9') && b[i] != '.' && b[i] != '-' {
// 			return false
// 		} else if (b[i] == '.' && i < len(b)-1 && b[i+1] == b[i]) || (b[i] == '-' && i < len(b)-1 && b[i+1] == b[i]) {
// 			return false
// 		}
// 	}
// 	return p
// }
