package src

import "net/http"

func DelBuc(w http.ResponseWriter, r *http.Request) {
	if bucname := r.PathValue("Bucket"); bucname == "Buckets.csv" {
		writeHttpError(w, http.StatusForbidden, "Invalid bucketnme", "You cannot delete")
	} else if mod, markeddel, e := writeTemp(Dir, bucname, "", "", true); e != nil { // give to kernel func
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Server dead")
	} else if !mod { // if not modified the metadata
		writeHttpError(w, http.StatusNotFound, "Not found bucket", "Check the bucket name")
	} else if markeddel { // modified to marked for deletion
		writeHttpError(w, http.StatusLocked, "bucket not empty", "bucket not empty")
	} else if e = writeHttpMessage(w, []byte("<deletedBucket><name>"+bucname+"</name></deletedBucket>")); e != nil {
		ErrPrint(e)
	}
}

func DelObj(w http.ResponseWriter, r *http.Request) {
	if bucname, objname := r.PathValue("Bucket"), r.PathValue("Object"); objname == "objects.csv" { // check the metadata's name
		writeHttpError(w, http.StatusForbidden, "Invalid object name", "you cannot delete this object")
	} else if moded, _, e := writeTemp(Dir+"/"+bucname, objname, "", "", true); e != nil { // delete entry of metadata
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Error with following reason")
	} else if !moded {
		writeHttpError(w, http.StatusNotFound, "Not found object", "In metadata not found")
	} else if mod, _, e := writeTemp(Dir, bucname, "", "", false); e != nil { // change the time of buckets.csv
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Fatal ERROR deleting object buckets temp")
	} else if !mod {
		writeHttpError(w, http.StatusInternalServerError, "bucket's time not modifiyed", "Fatal ERROR deleting object")
	} else if e = writeHttpMessage(w, []byte("<deletedobject><name>"+objname+"</name><bucket>"+bucname+"</bucket></deletedobject>")); e != nil {
		ErrPrint(e)
	}
}
