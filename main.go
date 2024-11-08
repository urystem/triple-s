package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"triple-s/src"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		src.Help()
	} else if port, e := func() (string, error) { // fuction CheckAndSetFlags
		if l := byte(len(os.Args)); l > 7 { // max len of args is 5
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
		} else if f, err := os.OpenFile(src.Dir+"/buckets.csv", os.O_CREATE|os.O_RDWR, 0o644); err != nil { // create or open bucket.csv file in this dir
			src.ErrPrint(err)
		} else {
			if finfo, e := f.Stat(); e != nil { // get info about bucket.csv file
				f.Close()
				src.ErrPrint(e)
			} else if finfo.Size() == 0 { // if it is empty file or just created file, write []Buchead for first line
				if _, e = f.WriteString(strings.Join(src.Buchead[:], ",") + "\n"); e != nil {
					f.Close()
					src.ErrPrint(e)
				}
				os.Stdout.WriteString("metadata created\n")
			} else {
				r := csv.NewReader(f)
				if e := src.Headchecker(&r, true); e != nil {
					f.Close()
					src.ErrPrint(e)
				}
			}
			f.Close()
			// if entries, e := os.ReadDir(dir); e != nil {
			// 	src.ErrPrint(e)
			// } else if len(entries) != 1 {
			// 	os.Stdout.WriteString("But in the this dir has another file(s) or directorie(s)\nBe carefully!")
			// }
			mux := http.NewServeMux()
			mux.HandleFunc("GET /{Bucket}", src.GetBuc)
			mux.HandleFunc("GET /", src.GetBucets)
			mux.HandleFunc("PUT /{Bucket}/{Object}", src.PutObj)
			mux.HandleFunc("PUT /{Bucket}", src.Putbuc)
			mux.HandleFunc("DELETE /{Bucket}/{Object}", src.DelObj)
			mux.HandleFunc("DELETE /{Bucket}", src.DelBuc)
			fmt.Println(port, src.Dir)
			fmt.Println("Start Server")
			if e = http.ListenAndServe(port, mux); e != nil {
				src.ErrPrint(e)
			}
		}
	}
}

// checker first line

// func for get method
func gettri(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Path
	fmt.Println(u, w)
	fmt.Println("GET")
}

func putobj(w http.ResponseWriter, r *http.Request) {
}

func puttri(w http.ResponseWriter, r *http.Request) {
	u := strings.Split(r.URL.Path, "/")
	if len(u[0]) == 0 { // if first element empty del first one
		u = u[1:]
	}
	if len(u[len(u)-1]) == 0 { // if last elemet empty del last one
		u = u[:len(u)-1]
	}
	switch len(u) {
	case 2:
		if f, e := os.Stat(src.Dir + "/" + u[0]); e != nil {
			http.Error(w, "", http.StatusInternalServerError)
		} else if !f.IsDir() {
			http.Error(w, "", http.StatusInternalServerError)
		} else if f, e := os.Create(src.Dir + "/" + u[0] + "/" + u[1]); e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		} else {
			defer r.Body.Close()
			if b, e := io.ReadAll(r.Body); e != nil {
				http.Error(w, e.Error(), http.StatusBadRequest)
			} else {
				if _, e := f.Write(b); e != nil {
					http.Error(w, e.Error(), http.StatusBadRequest)
				}
				f.Close()
			}
		}
	default:
		http.Error(w, "404 Not Found: The requested location is not allowed", http.StatusNotFound)
	}
}

// func for del method
func deltri(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Path
	b := r.Body
	fmt.Println(b, w)
	fmt.Println(u)
}

// func hadle and check methods and give each method func
func handls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		gettri(w, r)
	case "PUT":
		puttri(w, r)
	case "DELETE":
		deltri(w, r)
	default:
		http.Error(w, r.Method+" method not allowed", http.StatusMethodNotAllowed)
	}
}
