package src

import (
	"encoding/csv"
	"io"
	"net/http"
	"os"
)

func GetBucets(w http.ResponseWriter, r *http.Request) {
	if e := checkmeta(Dir+"/"+"buckets.csv", true); e != nil { // check metadata
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Error with metadata")
	} else if f, e := os.Open(Dir + "/" + "buckets.csv"); e != nil { // open metadata for read
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Error with metadata")
	} else {
		defer f.Close()
		reader := csv.NewReader(f)
		if e = Headchecker(&reader, true); e != nil { // check 2th time the header
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal error headbuc when 2th time checking")
		} else if e = writeHttpMessage(w, http.StatusOK, []byte("<buckets>")); e != nil {
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
	if _, e := os.Stat(Dir + "/" + bucname); e != nil { // info about this one bucket
		if os.IsNotExist(e) {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "bucket not found")
		} else {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "unknown error")
		}
	} else if e := checkmeta(Dir+"/"+bucname+"/"+"objects.csv", false); e != nil { // check metadata
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "checking metadata")
	} else if f, e := os.Open(Dir + "/" + bucname + "/" + "objects.csv"); e != nil { // open for read
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "opening metadata")
	} else {
		defer f.Close()
		read := csv.NewReader(f)
		if e = Headchecker(&read, false); e != nil { // check header 2th time
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal error headobj when 2th time checking")
		} else if e = writeHttpMessage(w, http.StatusOK, []byte("<objects>")); e != nil {
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
	if _, e := os.Stat(Dir + "/" + bucname); e != nil { // first of all, check the bucket exits or not
		if os.IsNotExist(e) {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "bucket not found")
		} else {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "unknown error")
		}
	} else if _, e = os.Stat(Dir + "/" + bucname + "/" + objname); e != nil { // check object file exitst or not
		if os.IsNotExist(e) {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "object not found")
		} else {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "unknown error")
		}
	} else if typ, err := checkHasAndreturntype(Dir+"/"+bucname+"/"+"objects.csv", objname, false); err != nil { // check and find the type of object
		writeHttpError(w, http.StatusInternalServerError, err.Error(), "metadata error")
	} else if f, e := os.Open(Dir + "/" + bucname + "/" + objname); e != nil { // open the object file for give to client
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "cannot open the object")
	} else {
		w.Header().Set("Content-Type", typ)  // set the type of file
		if _, e := io.Copy(w, f); e != nil { // then give
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "w,f error")
		}
	}
}
