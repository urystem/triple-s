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

func ErrPrint(e error) {
	os.Stdout.WriteString(e.Error() + "\n\n")
	os.Exit(1)
}

func Headchecker(r **csv.Reader, isBuc bool) error {
	if first, e := (*r).Read(); e != nil {
		return e
	} else if ch, bucOrObj := objhead, "object"; len(first) != 4 {
		return errors.New("in csv file's struct wrong len")
	} else {
		if isBuc {
			ch, bucOrObj = Buchead, "bucket"
		}
		for i, v := range first {
			if v != ch[i] {
				return errors.New(bucOrObj + ".csv file changed")
			}
		}
	}
	return nil
}

func writeHttpError(w http.ResponseWriter, code int, errorCode string, message string) {
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(code)
	if _, e := w.Write([]byte("<error><code>" + errorCode + "</code><message>" + message + "</message></error>")); e != nil {
		ErrPrint(e)
	}
}

func writeHttpMessage(w http.ResponseWriter, code int, by []byte) error {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "text/xml")
	_, e := w.Write(by)
	return e
}

func writeTemp(pathfile, bucOrObjname, size, con string, del bool) (bool, bool, error) { // univesal func, like a kernel of metadates)
	filename, tfilename, header := "objects.csv", "Objects.csv", &objhead
	if pathfile == Dir {
		filename, tfilename, header = "buckets.csv", "Buckets.csv", &Buchead
	}
	tf, er := os.Create(pathfile + "/" + tfilename) // creating temp
	if er != nil {
		return false, false, er
	}
	defer tf.Close()
	fcsv, err := os.Open(pathfile + "/" + filename) // open the original file only for reading
	if err != nil {
		return false, false, err
	}
	defer fcsv.Close()
	reader := csv.NewReader(fcsv)
	if e := Headchecker(&reader, pathfile == Dir); e != nil {
		return false, false, e
	}
	if _, e := tf.WriteString(strings.Join((*header)[:], ",") + "\n"); e != nil { // write the header to temp
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
					if files, err := os.ReadDir(pathfile + "/" + bucOrObjname); err != nil { // this this fatal error
						return false, false, err
					} else if len(files) == 1 && files[0].Name() == "objects.csv" && !files[0].IsDir() { // if there only 1 file it is to removing
						if err = os.RemoveAll(pathfile + "/" + bucOrObjname); err != nil { // removeall, because in it is will be dir with metadata objects.csv
							return false, false, err
						}
						continue // skip write to the metadata, it is for the buckets.csv
					} else {
						marketdel, fls[3] = true, "marked for deletion" // if in the dir not only the metadata objects.csv, for this one the function return one bool
					}
				} else if e := os.Remove(pathfile + "/" + bucOrObjname); e != nil { // remove/delete the object file
					return false, false, e
				} else {
					continue // skip for not write to the metadata, it is will be on after object deleted
				}
			} else if pathfile == Dir { // not deletion and it is Dir
				if fls[3] == "marked for deletion" { // Change to Active, because it is for update buckets.csv
					fls[3] = "Active"
				}
				fls[2] = time.Now().Format("2006-01-02T15:04:05") // Update time for buckets.csv
			} else { // Else that is for objects.csv
				fls[1], fls[2], fls[3] = size, con, time.Now().Format("2006-01-02T15:04:05") // update object's metadates
			}
		}
		if _, err := tf.WriteString(strings.Join(fls, ",") + "\n"); err != nil { // writing updated or not updated lines to temp
			return false, false, err
		}
	}
	if !was { // if not modified then just remove the temp
		return was, marketdel, os.Remove(pathfile + "/" + tfilename)
	}
	return was, marketdel, func() error { // if modified then remove original and rename the temp to original
		if er := os.Remove(pathfile + "/" + filename); er != nil {
			return er
		}
		return os.Rename(pathfile+"/"+tfilename, pathfile+"/"+filename)
	}()
}

func checkmeta(pathtofile string, isBuc bool) error {
	if f, e := os.Open(pathtofile); e != nil { // open for reading only
		return e
	} else {
		defer f.Close()
		read := csv.NewReader(f)
		if e := Headchecker(&read, isBuc); e != nil { // check the header
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

func checkHasAndreturntype(pathfile, objOrBucname string, isBuc bool) (string, error) {
	if f, e := os.Open(pathfile); e != nil {
		return "", e
	} else {
		defer f.Close()
		read := csv.NewReader(f)
		if e := Headchecker(&read, isBuc); e != nil {
			return "", e
		}
		for {
			if fls, er := read.Read(); er == io.EOF {
				break
			} else if er != nil {
				return "", er
			} else if fls[0] == objOrBucname {
				if isBuc { // if it is bucket, then must not been here
					return "", errors.New("bucket already exits in metadata")
				}
				return fls[2], nil // if it's object, it is right, we need the content type only
			}
		}
		if isBuc { // if bucket not found here, it is right, it for creating for new bucket
			return "", nil
		}
		return "", errors.New(objOrBucname + "not found") // if object not here, it is error, contype not found
	}
}
