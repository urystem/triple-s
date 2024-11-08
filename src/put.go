package src

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func Putbuc(w http.ResponseWriter, r *http.Request) {
	if fatalError != nil {
		ErrPrint(fatalError)
	} else if bucname := r.PathValue("Bucket"); bucname == "Buckets.csv" {
		writeHttpError(w, http.StatusBadRequest, "Invalid bucketnme", "You cannot create with this name")
	} else if !regexp.MustCompile(`^[-.a-z\d]{3,63}$`).MatchString(bucname) || regexp.MustCompile(`^\d+.\d+.\d+.\d+$`).MatchString(bucname) || regexp.MustCompile("^[.-]").MatchString(bucname) ||
		regexp.MustCompile("[.-]$").MatchString(bucname) || regexp.MustCompile(`\.\.`).MatchString(bucname) || regexp.MustCompile("--").MatchString(bucname) {
		writeHttpError(w, http.StatusBadRequest, "Invalid bucketname", "Check the name of bucket")
	} else if e := os.Mkdir(Dir+"/"+bucname, 0o755); e != nil {
		if os.IsExist(e) {
			writeHttpError(w, http.StatusConflict, e.Error(), "bucket already exist")
		} else {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "uknown error: cannot add bucket")
		}
	} else if e = os.WriteFile(Dir+"/"+bucname+"/objects.csv", []byte(strings.Join(objhead[:], ",")+"\n"), 0o644); e != nil {
		writeHttpError(w, http.StatusBadRequest, e.Error(), "can't write metadata to your bucket, From the next request will be turn off the server")
		fatalError = e
	} else if f, e := os.OpenFile(Dir+"/buckets.csv", os.O_RDWR, 0o644); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal Error with metadata: from next request will be turn of server")
		fatalError = e
	} else {
		defer f.Close()
		if err, ntime := func() error {
			r := csv.NewReader(f)
			if e = Headchecker(&r, true); e != nil {
				return e
			}
			for {
				if b, er := r.Read(); er == io.EOF {
					break
				} else if er != nil {
					return er
				} else if strings.TrimSpace(b[0]) == bucname {
					return errors.New(bucname + " bucket already exist in metadata")
				}
			}
			return nil
		}(), time.Now().Format("2006-01-02T15:04:05"); err != nil {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "Error metadata, from next request will be turn off the server")
			fatalError = e
		} else if _, e := f.WriteString(bucname + "," + ntime + "," + ntime + ",Active\n"); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "cannot write metadata:  From next request will be turn of server")
			fatalError = e
		} else if _, e = w.Write([]byte(xmlheader + "<bucketCreated><name>" + bucname + "</name>" + "<creationtime>" + ntime + "</creationtime>" + "</bucketCreated>")); e != nil {
			ErrPrint(e)
		}
	}
}

func PutObj(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if bucname := r.PathValue("Bucket"); fatalError != nil {
		ErrPrint(fatalError)
	} else if objname := r.PathValue("Object"); objname == "objects.csv" || objname == "Objects.csv" {
		writeHttpError(w, http.StatusBadRequest, "Invalid objectname", "You cannot create object that name")
	} else if f, e := os.Stat(Dir + "/" + bucname); e != nil {
		if os.IsNotExist(e) {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "There no that bucket")
		} else {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "error with bucket")
			fatalError = e
		}
	} else if consize, contype := r.Header.Get("Content-Length"), r.Header.Get("Content-Type"); !f.IsDir() {
		writeHttpError(w, http.StatusInternalServerError, "there are has file with this bucket name", "There are "+bucname+" is file not dir")
		fatalError = e
	} else if fobj, e := os.Create(Dir + "/" + bucname + "/" + objname); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Cannot create or update object")
	} else {
		defer fobj.Close()
		ntime := time.Now().Format("2006-01-02T15:04:05")
		if _, er := io.Copy(fobj, r.Body); er != nil {
			writeHttpError(w, http.StatusInternalServerError, er.Error(), "ERROR with reading")
			fatalError = er
		} else if moded, er := writeTemp(Dir+"/"+bucname, objname, consize, contype, false); er != nil {
			writeHttpError(w, http.StatusInternalServerError, er.Error(), "Fatal ERROR with temp")
			fatalError = er
		} else if !moded {
			os.Remove(Dir + "/" + bucname + "/" + "Objects.csv")
			if file, err := os.OpenFile(Dir+"/"+bucname+"/"+"objects.csv", os.O_WRONLY|os.O_APPEND, 0o644); err != nil {
				writeHttpError(w, http.StatusInternalServerError, err.Error(), "Error opening metadata to add")
				fatalError = err
			} else if _, erre := file.WriteString(objname + "," + consize + "," + contype + "," + ntime + "\n"); erre != nil {
				file.Close()
				writeHttpError(w, http.StatusInternalServerError, erre.Error(), "Cannot add matadata")
				fatalError = erre
			} else {
				file.Close()
			}
		} else if e := temptocsv(Dir + "/" + bucname); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatality temp to base")
			fatalError = e
		}
		if fatalError == nil {
			if mod, er := writeTemp(Dir, bucname, "", "", false); er != nil {
				writeHttpError(w, http.StatusInternalServerError, er.Error(), "Fatality temp to base")
				fatalError = e
			} else if !mod {
				writeHttpError(w, http.StatusInternalServerError, "Metadata error", "not fuond in metadata bucket")
			} else if e = temptocsv(Dir); e != nil {
				writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatality temp to base bucket csv")
				fatalError = e
			} else if _, e := w.Write([]byte(xmlheader + "<" + bucname + ">" + "<putobject>" + "<name>" + objname + "</name>" + "<size>" + consize + "</size>" +
				"<type>" + contype + "</type>" + "<lastModified>" + ntime + "</lastModified>" + "</putobject>" + "</" + bucname + ">")); e != nil {
				ErrPrint(e)
			}
		}
	}
}
