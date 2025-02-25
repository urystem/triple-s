package main

import (
	"encoding/csv"
	"flag"
	"net/http"
	"os"
	"regexp"
	"strings"

	"triple-s/src"
)

func main() {
	port, dir := flag.String("port", "8080", "set port"), flag.String("dir", "data", "directory")
	flag.Usage = func() {
		os.Stdout.WriteString(`Simple Storage Service.
**Usage:**
	triple-s [-port <N>] [-dir <S>]
	triple-s --help
		
	**Options:**
- --help     Show this screen.
- --port N   Port number (default is 8080).
- --dir S    Path to the directory (default is to the 'data' directory).` + "\n")
	}
	flag.Parse()
	if *port = strings.TrimLeft(*port, "0 "); !regexp.MustCompile(`^\d+$`).MatchString(*port) {
		os.Stdout.WriteString("Invalid port\n")
	} else if dirinfo, err := os.Stat(*dir); err != nil {
		os.Stdout.WriteString("diretory error: " + err.Error() + "\n")
	} else if !dirinfo.IsDir() {
		os.Stdout.WriteString("In the path was file, not directory\n")
	} else if err := func() error {
		src.Dir = *dir
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
	} else {
		mux := http.NewServeMux()
		mux.HandleFunc("PUT /{Bucket}", src.Putbuc)
		mux.HandleFunc("PUT /{Bucket}/{Object}", src.PutObj)
		mux.HandleFunc("DELETE /{Bucket}", src.DelBuc)
		mux.HandleFunc("DELETE /{Bucket}/{Object}", src.DelObj)
		mux.HandleFunc("GET /{Bucket}/{Object}", src.GetObj)
		mux.HandleFunc("GET /", src.GetBucets)
		mux.HandleFunc("GET /{Bucket}/", src.GetBuc)
		os.Stdout.WriteString("Port: " + *port + "\tDir: " + src.Dir + "\nServer starting\n")
		src.ErrPrint(http.ListenAndServe(":"+*port, mux))
	}
}
