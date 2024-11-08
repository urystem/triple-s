package src

import (
	"net/http"
)

func DelBuc(w http.ResponseWriter, r *http.Request) {
	if fatalError != nil {
		ErrPrint(fatalError)
	} else if bucname := r.PathValue("Bucket"); bucname == "Buckets.csv" {
		writeHttpError(w, http.StatusBadRequest, "Invalid bucketnme", "You cannot delete")
	} else if mod, e := writeTemp(Dir, bucname, "", "", true); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Server dead")
		fatalError = e
	} else if !mod {
		writeHttpError(w, http.StatusBadRequest, "Not found bucket", "Check the bucket name")
	} else if e = temptocsv(Dir); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal ERROR")
		fatalError = e
	} else if _, e = w.Write([]byte(xmlheader + "<deletedBucket>" + "<name>" + bucname + "</name>" + "</deletedBucket>")); e != nil {
		writeHttpError(w, http.StatusBadGateway, e.Error(), "Fatality del")
		fatalError = e
	}
}

func DelObj(w http.ResponseWriter, r *http.Request) {
	if fatalError != nil {
		ErrPrint(fatalError)
	} else if bucname, objname := r.PathValue("Bucket"), r.PathValue("Object"); objname == "objects.csv" {
		writeHttpError(w, http.StatusBadRequest, "Invalid object name", "you cannot delete this object")
	} else if moded, e := writeTemp(Dir+"/"+bucname, objname, "", "", true); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal ERROR deleting object")
		fatalError = e
	} else if !moded {
		writeHttpError(w, http.StatusBadRequest, "Not found object", "Check the object name")
	} else if e = temptocsv(Dir + "/" + bucname); e != nil {
		writeHttpError(w, http.StatusBadRequest, e.Error(), "temptocsv error")
		fatalError = e
	} else if _, e = w.Write([]byte(xmlheader + "<deletedobject>" + "<name>" + objname + "</name>" + "<bucket>" + bucname + "</bucket>" + "</deletedobject>")); e != nil {
		writeHttpError(w, http.StatusBadGateway, e.Error(), "fatal error with message")
		fatalError = e
	}
}
