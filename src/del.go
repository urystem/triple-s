package src

import (
	"net/http"
	"os"
)

func DelBuc(w http.ResponseWriter, r *http.Request) {
	if bucname := r.PathValue("Bucket"); bucname == "Buckets.csv" {
		writeHttpError(w, http.StatusBadRequest, "Invalid bucketnme", "You cannot delete")
	} else if mod, marketdel, e := writeTemp(Dir, bucname, "", "", true); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Server dead")
	} else if !mod {
		if e = os.Remove(Dir + "/" + "Buckets.csv"); e != nil {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "Cannot remove bucket temp metadata")
		}
		writeHttpError(w, http.StatusBadRequest, "Not found bucket", "Check the bucket name")
	} else if e = temptocsv(Dir); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal ERROR")
	} else if marketdel {
		writeHttpError(w, http.StatusInternalServerError, "bucket not empty", "bucket not empty")
	} else if e = writeHttpMessage(w, []byte("<deletedBucket>"+"<name>"+bucname+"</name>"+"</deletedBucket>")); e != nil {
		ErrPrint(e)
	}
}

func DelObj(w http.ResponseWriter, r *http.Request) {
	if bucname, objname := r.PathValue("Bucket"), r.PathValue("Object"); objname == "objects.csv" {
		writeHttpError(w, http.StatusBadRequest, "Invalid object name", "you cannot delete this object")
	} else if moded, _, e := writeTemp(Dir+"/"+bucname, objname, "", "", true); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal ERROR deleting object object temp")
	} else if !moded {
		if e = os.Remove(Dir + "/" + bucname + "/" + "Objects.csv"); e != nil {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "Cannot remove object temp metadata")
		}
		writeHttpError(w, http.StatusBadRequest, "Not found object", "Check the object name")
	} else if e = temptocsv(Dir + "/" + bucname); e != nil {
		writeHttpError(w, http.StatusBadRequest, e.Error(), "temptocsv error")
	} else if mod, _, e := writeTemp(Dir, bucname, "", "", false); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal ERROR deleting object buckets temp")
	} else if !mod {
		writeHttpError(w, http.StatusInternalServerError, "bucket not modifiyed", "Fatal ERROR deleting object buckets temp "+bucname+" not found")
	} else if e = temptocsv(Dir); e != nil {
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal ERROR temp to original bucket")
	} else if e = writeHttpMessage(w, []byte("<deletedobject><name>"+objname+"</name><bucket>"+bucname+"</bucket></deletedobject>")); e != nil {
		ErrPrint(e)
	}
}
