package src

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetBucets(w http.ResponseWriter, r *http.Request) {
	if e := checkmeta(Dir+"/"+"buckets.csv", true); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Error with metadata")
	} else if f, e := os.Open(Dir + "/" + "buckets.csv"); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Error with metadata")
		fatalError = e
	} else {
		defer f.Close()
		reader := csv.NewReader(f)
		if e = Headchecker(&reader, true); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal error headbuc")
			fatalError = e
		} else if _, e = w.Write([]byte(xmlheader + "<buckets>")); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal error writing message")
			fatalError = e
		} else if er := getprinter(&reader, w, true); er != nil {
			ErrPrint(er)
		} else if _, e = w.Write([]byte("</buckets>")); e != nil {
			ErrPrint(e)
		}
	}
}

func GetBuc(w http.ResponseWriter, r *http.Request) {
	bucname := r.PathValue("Bucket")
	fmt.Println(bucname)
	if fn, e := os.Stat(Dir + "/" + bucname); e != nil {
		if os.IsNotExist(e) {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "not found bucket")
		} else {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "unknown error bucket")
			fatalError = e
		}
	} else if !fn.IsDir() {
		writeHttpError(w, http.StatusInternalServerError, "it is file", "is not dir")
		fatalError = e
	} else if e = checkmeta(Dir+"/"+bucname+"/"+"objects.csv", false); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "checking metadata")
		fatalError = e
	} else if fn, e := os.Open(Dir + "/" + "buckets.csv"); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "opening metadata")
		fatalError = e
	} else {
		defer fn.Close()
		read := csv.NewReader(fn)
		if e = Headchecker(&read, false); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "second check header")
			fatalError = e
		} else if _, e = w.Write([]byte(xmlheader + "<bucket><name>" + bucname + "</name>")); e != nil {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "writing")
			fatalError = e
		} else if e = getprinter(&read, w, false); e != nil {
			ErrPrint(e)
		} else if _, e = w.Write([]byte("</bucket>")); e != nil {
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
