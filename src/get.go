package src

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"os"
)

func GetBucets(w http.ResponseWriter, r *http.Request) {
	if e := checkmeta(Dir+"/"+"buckets.csv", true); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Error with metadata")
	} else if f, e := os.Open(Dir + "/" + "buckets.csv"); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Error with metadata")
	} else {
		defer f.Close()
		reader := csv.NewReader(f)
		if e = Headchecker(&reader, true); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal error headbuc")
		} else if e = writeHttpMessage(w, []byte("<buckets>")); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal error writing message")
		} else if er := getprinter(&reader, w, true); er != nil {
			ErrPrint(er)
		} else if _, e = w.Write([]byte("</buckets>")); e != nil {
			ErrPrint(e)
		}
	}
}

func GetBuc(w http.ResponseWriter, r *http.Request) {
	bucname := r.PathValue("Bucket")
	if _, e := os.Stat(Dir + "/" + bucname); e != nil {
		if os.IsNotExist(e) {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "bucket not found")
		} else {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "unknown error")
		}
	} else if e := checkmeta(Dir+"/"+bucname+"/"+"objects.csv", false); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "checking metadata")
	} else if f, e := os.Open(Dir + "/" + bucname + "/" + "objects.csv"); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "opening metadata")
	} else {
		defer f.Close()
		read := csv.NewReader(f)
		if e = Headchecker(&read, false); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "second check header")
		} else if e = writeHttpMessage(w, []byte("<objects>")); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "writing")
		} else if e = getprinter(&read, w, false); e != nil {
			ErrPrint(e)
		} else if _, e = w.Write([]byte("</objects>")); e != nil {
			ErrPrint(e)
		}
	}
}

func getprinter(r **csv.Reader, w http.ResponseWriter, isBuc bool) error {
	var xmlf [5]string
	if isBuc {
		xmlf[0], xmlf[1], xmlf[2], xmlf[3], xmlf[4] = "<bucket><name>", "</name><creationTime>", "</creationTime><LastModifiedTime>", "</LastModifiedTime><status>", "</status></bucket>"
	} else {
		xmlf[0], xmlf[1], xmlf[2], xmlf[3], xmlf[4] = "<object><objectKey>", "</objectKey><size>", "</size><contentType>", "</contentType><lastModified>", "</lastModified></object>"
	}
	for {
		if rec, e := (*r).Read(); e == io.EOF {
			break
		} else if e != nil {
			return e
		} else if _, e = w.Write([]byte(xmlf[0] + rec[0] + xmlf[1] + rec[1] + xmlf[2] + rec[2] + xmlf[3] + rec[3] + xmlf[4])); e != nil {
			return e
		}
	}
	return nil
}

func GetObj(w http.ResponseWriter, r *http.Request) {
	bucname, objname := r.PathValue("Bucket"), r.PathValue("Object")
	if _, e := os.Stat(Dir + "/" + bucname); e != nil {
		if os.IsNotExist(e) {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "bucket not found")
		} else {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "unknown error")
		}
	} else if _, e = os.Stat(Dir + "/" + bucname + "/" + objname); e != nil {
		if os.IsNotExist(e) {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "n=")
		} else {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "unknown error")
		}
	} else if conty, err := func() (string, error) {
		if f, e := os.Open(Dir + "/" + bucname + "/" + "objects.csv"); e != nil {
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
				} else if fls[0] == objname {
					return fls[2], nil
				}
			}
			return "", errors.New("bucket: " + bucname + " object: " + objname + "Object not found in metadata")
		}
	}(); err != nil {
		writeHttpError(w, http.StatusInternalServerError, err.Error(), "metadata error")
	} else if f, e := os.Open(Dir + "/" + bucname + "/" + objname); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "cannot open the object")
	} else {
		w.Header().Set("Content-Type", conty)
		if _, e := io.Copy(w, f); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "w,f error")
		}
	}
}
