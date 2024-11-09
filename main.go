package main

import (
	"encoding/csv"
	"errors"
	"net/http"
	"os"
	"strings"

	"triple-s/src"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		src.Help()
	} else if port, e := func() (string, error) { // fuction CheckAndSetFlags
		if l := byte(len(os.Args)); l > 5 { // max len of args is 5
			return "", errors.New("too many args")
		} else if l%2 == 0 { // even count args is invalid
			return "", errors.New("invalid counts of args")
		} else {
			var d, p bool // for was dir or port before
			var index byte
			os.Args[0] = ""                   // by default s,n 0 ,so that is way i need change to nothiong
			for i := byte(1); i < l; i += 2 { // for every even elements,for example 1 is even because it is index)
				switch os.Args[i] {
				case "-dir", "--dir": // if it is flag dir
					if d { // was or not dir flag before
						return "", errors.New("repeated dir flag")
					}
					d = !d                      // set d to yes that true, s index to next index of i that i+1
					if len(os.Args[i+1]) == 0 { // if dir flag empty
						return "", errors.New("empty dir")
					}
					src.Dir = os.Args[i+1]
				case "-port", "--port": // if it is port flag
					if p {
						return "", errors.New("repeated port flag") // was or not port flag before
					}
					index, p = i+1, !p // set boolean and next index that is port flag
					if func() bool {   // check for 0 port, it is auto port
						for _, v := range os.Args[index] {
							if v != '0' {
								return false
							}
						}
						return true
					}() {
						return "", errors.New("empty or 0 port")
					}
				default:
					return "", errors.New("unknown flag - " + os.Args[i]) // if unknown for even flags
				}
			}
			return ":" + os.Args[index], nil // last legal return
		}
	}(); e != nil { // that func's return error?
		src.ErrPrint(e)
	} else {
		if len(port) == 1 { // it cannot be 0, because i will add to return ':'
			port = ":8080"
		}
		if len(src.Dir) == 0 { // if dir it is arg[0] that empty, set to default path
			src.Dir = "data"
		}
		if ds, e := os.Stat(src.Dir); e != nil { // getting info about this path
			src.ErrPrint(e)
		} else if !ds.IsDir() { // if it's not dir, folder
			src.ErrPrint(errors.New("it is not dir"))
		} else if err := func() error {
			if f, e := os.OpenFile(src.Dir+"/buckets.csv", os.O_CREATE|os.O_RDWR, 0o644); e != nil {
				return e
			} else {
				defer f.Close()
				if finfo, e := f.Stat(); e != nil { // get info about bucket.csv file
					return e
				} else if finfo.Size() == 0 { // if it is empty file or just created file, write []Buchead for first line
					_, e = f.WriteString(strings.Join(src.Buchead[:], ",") + "\n")
					return e
				}
				r := csv.NewReader(f)
				return src.Headchecker(&r, true)
			}
		}(); err != nil {
			src.ErrPrint(err)
		}
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /{Bucket}", src.Putbuc)
		mux.HandleFunc("PUT /{Bucket}/{Object}", src.PutObj)
		mux.HandleFunc("DELETE /{Bucket}", src.DelBuc)
		mux.HandleFunc("DELETE /{Bucket}/{Object}", src.DelObj)
		mux.HandleFunc("GET /{Bucket}/{Object}", src.GetObj)
		mux.HandleFunc("GET /", src.GetBucets)
		mux.HandleFunc("GET /{Bucket}/", src.GetBuc)
		os.Stdout.WriteString("Port: " + port + "\tDir: " + src.Dir + "\nServer starting\n")
		if e = http.ListenAndServe(port, mux); e != nil {
			src.ErrPrint(e)
		}
	}
}
