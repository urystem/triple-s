package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func help() {
	os.Stdout.WriteString(`Simple Storage Service.
**Usage:**
	triple-s [-port <N>] [-dir <S>]
	triple-s --help

**Options:**
- --help     Show this screen.
- --port N   Port number (default is 8080).
- --dir S    Path to the directory (default is current directory).` + "\n")
}

func errPrint(e error) {
	os.Stdout.WriteString(e.Error() + "\n\n")
	help()
	os.Exit(1)
}

var (
	Dir     string
	Buchead [4]string = [4]string{"Name", "CreationTime", "LastModifiedTime", "Status"}
	Object  [4]string = [4]string{"ObjectKey", "Size", "ContentType", "LastModified"}
	BucFile **os.File
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		help()
	} else if dir, port, e := func() (string, string, error) { // fuction CheckAndSetFlags
		if l := byte(len(os.Args)); l > 5 { // max len of args is 5
			return "", "", errors.New("too many args")
		} else if l%2 == 0 { // even count args is invalid
			return "", "", errors.New("invalid counts of args")
		} else {
			var (
				d, p bool // for was dir or port before
				s, n byte // for remember indexs of args
			)
			os.Args[0] = ""                   // by default s,n 0 ,so that is way i need change to nothiong
			for i := byte(1); i < l; i += 2 { // for every even elements,for example 1 is even because it is index)
				switch os.Args[i] {
				case "-dir", "--dir": // if it is flag dir
					if d { // was or not dir flag before
						return "", "", errors.New("repeated dir flag")
					}
					d, s = !d, i+1            // set d to yes that true, s index to next index of i that i+1
					if len(os.Args[s]) == 0 { // if dir flag empty
						return "", "", errors.New("empty dir")
					}
				case "-port", "--port": // if it is port flag
					if p {
						return "", "", errors.New("repeated port flag") // was or not port flag before
					}
					p, n = !p, i+1   // set boolean and next index that is port flag
					if func() bool { // check for 0 port, it is auto port
						for _, v := range os.Args[n] {
							if v != '0' {
								return false
							}
						}
						return true
					}() {
						return "", "", errors.New("for port 0 you cannot find which port connected")
					}
				default:
					return "", "", errors.New("unknown flag - " + os.Args[i]) // if unknown for even flags
				}
			}
			return os.Args[s], ":" + os.Args[n], nil // last legal return
		}
	}(); e != nil { // that func's return error?
		errPrint(e)
	} else {
		if len(port) == 1 { // it cannot be 0, because i will add to return ':'
			port = ":8080"
		}
		if len(dir) == 0 { // if dir it is arg[0] that empty, set to default path
			dir = "data"
		}
		if ds, e := os.Stat(dir); e != nil { // getting info about this path
			errPrint(e)
		} else if !ds.IsDir() { // if it's not dir, folder
			errPrint(errors.New("it is not dir"))
		} else if f, err := os.OpenFile(dir+"/buckets.csv", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644); err != nil { // create or open bucket.csv file in this dir
			errPrint(err)
		} else {
			BucFile = &f
			defer f.Close()
			Dir = dir
			if finfo, e := f.Stat(); e != nil { // get info about bucket.csv file
				errPrint(e)
			} else if finfo.Size() == 0 { // if it is empty file or just created file, write []Buchead for first line
				if _, e = f.WriteString(strings.Join(Buchead[:], ",")); e != nil {
					errPrint(e)
				}
				// writer := csv.NewWriter(f)
				// writer.Write(Buchead[:])
				// writer.Flush()
			} else if first, e := csv.NewReader(f).Read(); e != nil { // else file no empty, read first line for []Buchead
				errPrint(e)
			} else if len(first) != 4 || !headchecker(first) { // check for []Buckead
				errPrint(errors.New("buckets.csv already exitsted, but there another strcuts"))
			}
			mux := http.NewServeMux()
			mux.HandleFunc("PUT /{Bucket}", puttri)
			
			/*

							http.HandleFunc("PUT /{BucketName}", func1) //mux.DefaulServeMux
							http.HandleFunc("GET /{BucketName}", func2)
							http.HandleFunc("GET /{BucketName}", func2)
							http.HandleFunc("PUT /{BucketName}/{obj}", func3)

							func func1(w http.RW, r *http.R){
								name:= r.PathValue(BucketName)
							}
							func func3(w http.RW, r *http.R){
								nameBucket:= r.PathValue(BucketName)
								obj := r.Pathvalue(obj)
							}
				http.HandleFunc("/", nullFunc)





							mux := http.NewServeMux()

				if e = http.ListenAndServe(port, mux); e != nil {
								errPrint(e)
							}

							www.NW.ru/admin/
							www.NW.ru/user/
							muxUser, muxAdmin
							mux = muxUser, muxAdmin
			*/
			fmt.Println(port, dir)
			fmt.Println("Start Server")
			if e = http.ListenAndServe(port, mux); e != nil {
				errPrint(e)
			}
		}
	}
}

// checker first line to []Bucked
func headchecker(a []string) bool {
	for i := range Buchead {
		if strings.TrimSpace(a[i]) != Buchead[i] {
			return false
		}
	}
	return true
}

// func for get method
func gettri(w http.ResponseWriter, r *http.Request) {
	u := r.URL.Path
	fmt.Println(u, w)
	fmt.Println("GET")
}

// func for put method
func puttri(w http.ResponseWriter, r *http.Request) {
	u := strings.Split(r.URL.Path, "/")
	if len(u[0]) == 0 { // if first element empty del first one
		u = u[1:]
	}
	if len(u[len(u)-1]) == 0 { // if last elemet empty del last one
		u = u[:len(u)-1]
	}
	switch len(u) {
	case 1: // if there are only 1 location, then it is for creation bucket
		if !regexp.MustCompile(`^[-.a-z\d]{3,63}$`).MatchString(u[0]) || regexp.MustCompile(`^\d+.\d+.\d+.\d+$`).MatchString(u[0]) || regexp.MustCompile("^[.-]").MatchString(u[0]) ||
			regexp.MustCompile("[.-]$").MatchString(u[0]) || regexp.MustCompile(`\.\.`).MatchString(u[0]) || regexp.MustCompile("--").MatchString(u[0]) {
			http.Error(w, "Check bucket name", http.StatusBadRequest)
		} else if e := os.Mkdir(Dir+"/"+u[0], 0o755); e != nil {
			http.Error(w, e.Error(), http.StatusConflict)
		} else if e = os.WriteFile(Dir+"/"+u[0]+"/objects.csv", []byte(strings.Join(Object[:], ",")), 0o644); e != nil {
			http.Error(w, e.Error(), http.StatusInternalServerError)
		} else {
			r, first := csv.NewReader(*BucFile), true
			for {
				if b, e := r.Read(); e == io.EOF || first {
					break
				} else if e != nil {
					http.Error(w, e.Error(), http.StatusInternalServerError)
				} else if len(b) != 4 {
					http.Error(w, "", http.StatusInternalServerError)
				} else if strings.TrimSpace(b[0]) == u[0] {
					http.Error(w, "confilct of csv file", http.StatusConflict)
				}
			}

		}
	case 2:
		if f, e := os.Stat(Dir + "/" + u[0]); e != nil {
			http.Error(w, "", http.StatusInternalServerError)
		} else if !f.IsDir() {
			http.Error(w, "", http.StatusInternalServerError)
		} else if f, e := os.Create(Dir + "/" + u[0] + "/" + u[1]); e != nil {
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
