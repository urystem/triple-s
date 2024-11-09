package src

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	Dir     string
	Buchead [4]string = [4]string{"Name", "CreationTime", "LastModifiedTime", "Status"}
	objhead [4]string = [4]string{"ObjectKey", "Size", "ContentType", "LastModified"}
)

func writeHttpError(w http.ResponseWriter, code int, errorCode string, message string) {
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(code)
	if _, e := w.Write([]byte("<error><code>" + errorCode + "</code><message>" + message + "</message></error>")); e != nil {
		ErrPrint(e)
	}
}

func writeHttpMessage(w http.ResponseWriter, by []byte) error {
	w.Header().Set("Content-Type", "text/xml")
	_, e := w.Write(by)
	return e
}

func Help() {
	os.Stdout.WriteString(`Simple Storage Service.
**Usage:**
	triple-s [-port <N>] [-dir <S>]
	triple-s --help

**Options:**
- --help     Show this screen.
- --port N   Port number (default is 8080).
- --dir S    Path to the directory (default is to the 'data' directory).` + "\n")
}

func ErrPrint(e error) {
	os.Stdout.WriteString(e.Error() + "\n\n")
	Help()
	os.Exit(1)
}

func Headchecker(r **csv.Reader, isBuc bool) error {
	if first, e := (*r).Read(); e != nil {
		return e
	} else if ch := objhead; len(first) != 4 {
		return errors.New("in csv file's struct wrong")
	} else {
		if isBuc {
			ch = Buchead
		}
		for i, v := range first {
			if f := "object"; v != ch[i] {
				if isBuc {
					f = "bucket"
				}
				return errors.New(f + ".csv file changed")
			}
		}
	}
	return nil
}

func writeTemp(pathfile, bucOrObjname, size, con string, del bool) (bool, bool, error) { // 0 FOR BUCKET DELETING, 1 FOR UPDATE OBJECT, 2 DEL OBJ
	filename, tfilename, header := "objects.csv", "Objects.csv", &objhead
	if pathfile == Dir {
		filename, tfilename, header = "buckets.csv", "Buckets.csv", &Buchead
	}
	tf, er := os.Create(pathfile + "/" + tfilename)
	if er != nil {
		return false, false, er
	}
	defer tf.Close()
	fcsv, err := os.Open(pathfile + "/" + filename)
	if err != nil {
		return false, false, err
	}
	defer fcsv.Close()
	reader := csv.NewReader(fcsv)
	if e := Headchecker(&reader, pathfile == Dir); e != nil {
		return false, false, e
	}
	if _, e := tf.WriteString(strings.Join((*header)[:], ",") + "\n"); e != nil {
		return false, false, e
	}
	var was, marketdel bool
	for {
		fls, er := reader.Read()
		if er == io.EOF {
			break
		} else if er != nil {
			return false, false, er
		} else if fls[0] == bucOrObjname { // for object.csv
			if was {
				return false, false, errors.New("repeated entry")
			}
			was = true
			if del {
				if pathfile == Dir { // del bucket
					if files, err := os.ReadDir(pathfile + "/" + bucOrObjname); err != nil {
						return false, false, err
					} else if len(files) == 1 && files[0].Name() == "objects.csv" && !files[0].IsDir() { // if there only 1 file it is to removing
						if err = os.RemoveAll(pathfile + "/" + bucOrObjname); err != nil {
							return false, false, err
						}
						continue
					} else {
						marketdel, fls[3] = true, "Market for deleting"
					}
				} else if e := os.Remove(pathfile + "/" + bucOrObjname); e != nil {
					return false, false, e
				} else {
					continue
				}
			} else if pathfile == Dir {
				if fls[3] == "Market for deleting" {
					fls[3] = "Active"
				}
				fls[2] = time.Now().Format("2006-01-02T15:04:05")
			} else {
				fls[1], fls[2], fls[3] = size, con, time.Now().Format("2006-01-02T15:04:05")
			}
		}
		if _, err := tf.WriteString(strings.Join(fls, ",") + "\n"); err != nil {
			return false, false, err
		}
	}
	return was, marketdel, nil
}

func temptocsv(path string) error {
	filename, tfilename := "objects.csv", "Objects.csv"
	if path == Dir {
		filename, tfilename = "buckets.csv", "Buckets.csv"
	}
	if er := os.Remove(path + "/" + filename); er != nil {
		return er
	}
	return os.Rename(path+"/"+tfilename, path+"/"+filename)
}

func checkmeta(pathtofile string, isBuc bool) error {
	if f, e := os.Open(pathtofile); e != nil {
		return e
	} else {
		defer f.Close()
		read := csv.NewReader(f)
		if e := Headchecker(&read, isBuc); e != nil {
			return e
		}
		for {
			if _, e := read.Read(); e == io.EOF {
				break
			} else if e != nil {
				return e
			}
		}
		return nil
	}
}

func checkAndreturntype(path, objOrBucname string) (string, error) {
	if path == Dir {
		path += "/buckets.csv"
	} else {
		path += "/objects.csv"
	}
	if f, e := os.Open(path); e != nil {
		return "", e
	} else {
		defer f.Close()
		read := csv.NewReader(f)
		if e := Headchecker(&read, false); e != nil {
			return "", e
		}
		for {
			if fls, er := read.Read(); er == io.EOF {
				break
			} else if er != nil {
				return "", er
			} else if fls[0] == objOrBucname {
				return fls[2], nil
			}
		}
		return "", errors.New(objOrBucname + "not found")
	}
}
