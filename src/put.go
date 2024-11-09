package src

import (
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func Putbuc(w http.ResponseWriter, r *http.Request) {
	if bucname := r.PathValue("Bucket"); bucname == "Buckets.csv" { // check to the temp file
		writeHttpError(w, http.StatusBadRequest, "Invalid bucketnme", "You cannot create with this name")
	} else if !regexp.MustCompile(`^[-.a-z\d]{3,63}$`).MatchString(bucname) || regexp.MustCompile(`^\d+.\d+.\d+.\d+$`).MatchString(bucname) || regexp.MustCompile("^[.-]").MatchString(bucname) ||
		regexp.MustCompile("[.-]$").MatchString(bucname) || regexp.MustCompile(`\.\.`).MatchString(bucname) || regexp.MustCompile("--").MatchString(bucname) { // rex check
		writeHttpError(w, http.StatusBadRequest, "invalid bucketname", "check the name of bucket")
	} else if _, err := checkHasAndreturntype(Dir+"/buckets.csv", bucname, true); err != nil { // check existed before in the metadata
		writeHttpError(w, http.StatusBadRequest, err.Error(), "found in metada with this name")
	} else if e := os.Mkdir(Dir+"/"+bucname, 0o755); e != nil { // try mkdir with this name
		if os.IsExist(e) {
			writeHttpError(w, http.StatusConflict, e.Error(), "found in metadata: bucket already exist")
		} else {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "uknown error: cannot create bucket")
		}
	} else if e = os.WriteFile(Dir+"/"+bucname+"/objects.csv", []byte(strings.Join(objhead[:], ",")+"\n"), 0o644); e != nil { // then into this bucket create file and write object header
		writeHttpError(w, http.StatusBadRequest, e.Error(), "can't create metadata and write header")
	} else if f, e := os.OpenFile(Dir+"/buckets.csv", os.O_APPEND|os.O_WRONLY, 0o644); e != nil { // open to add the new bucket
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "cannot open the metadata to add thw the bucket")
	} else {
		defer f.Close()
		ntime := time.Now().Format("2006-01-02T15:04:05")
		if _, e := f.WriteString(bucname + "," + ntime + "," + ntime + ",Active\n"); e != nil { // add the new bucket
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "cannot add the bucket to metadata")
		} else if e = writeHttpMessage(w, []byte("<bucketCreated><name>"+bucname+"</name><creationtime>"+ntime+"</creationtime></bucketCreated>")); e != nil { // message
			ErrPrint(e)
		}
	}
}

func PutObj(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if bucname, objname := r.PathValue("Bucket"), r.PathValue("Object"); objname == "objects.csv" || objname == "Objects.csv" { // check the name for temp file
		writeHttpError(w, http.StatusBadRequest, "Invalid objectname", "You cannot create object that name")
	} else if f, e := os.Stat(Dir + "/" + bucname); e != nil { // check the bucket exist or not
		if os.IsNotExist(e) {
			writeHttpError(w, http.StatusBadRequest, e.Error(), "There no that bucket")
		} else {
			writeHttpError(w, http.StatusInternalServerError, e.Error(), "unknown error with bucket")
		}
	} else if consize, contype := r.Header.Get("Content-Length"), r.Header.Get("Content-Type"); !f.IsDir() { // check the bucket is the directory
		writeHttpError(w, http.StatusInternalServerError, "there are has file with this bucket name", "There are "+bucname+" is file not dir")
	} else if fobj, e := os.Create(Dir + "/" + bucname + "/" + objname); e != nil { // create or update the object file
		writeHttpError(w, http.StatusInternalServerError, e.Error(), "Cannot create the new or update object")
	} else {
		defer fobj.Close()
		var waserr bool
		ntime := time.Now().Format("2006-01-02T15:04:05")
		if _, er := io.Copy(fobj, r.Body); er != nil { // read the body and write to the object file
			writeHttpError(w, http.StatusInternalServerError, er.Error(), "ERROR with copiny body to object")
			waserr = true
		} else if moded, _, er := writeTemp(Dir+"/"+bucname, objname, consize, contype, false); er != nil { // create the updated temp file
			writeHttpError(w, http.StatusInternalServerError, er.Error(), "Error with metadata about of objects")
			waserr = true
		} else if !moded { // if in the temp none change, just remove the temp and open the objects.csv to append the new entry
			if file, err := os.OpenFile(Dir+"/"+bucname+"/"+"objects.csv", os.O_WRONLY|os.O_APPEND, 0o644); err != nil {
				writeHttpError(w, http.StatusInternalServerError, err.Error(), "Error opening metadata to add the new object")
				waserr = true
			} else if _, erre := file.WriteString(objname + "," + consize + "," + contype + "," + ntime + "\n"); erre != nil {
				file.Close()
				writeHttpError(w, http.StatusInternalServerError, erre.Error(), "Cannot add to the matadata")
				waserr = true
			} else {
				file.Close()
			}
		}
		if !waserr { // this is pause important, because objects.csv may be changed or not in up⬆️😊
			if mod, _, er := writeTemp(Dir, bucname, "", "", false); er != nil { // then it is for update the buckets.csv
				writeHttpError(w, http.StatusInternalServerError, er.Error(), "error with updating the metadata")
			} else if !mod { // it is for buckets.csv, so this one must modified the last time
				writeHttpError(w, http.StatusInternalServerError, "Metadata error", "not found in metadata bucket")
			} else if e := writeHttpMessage(w, []byte("<"+bucname+">"+"<putobject><name>"+objname+"</name><size>"+
				consize+"</size><type>"+contype+"</type><lastModified>"+ntime+"</lastModified></putobject></"+bucname+">")); e != nil {
				ErrPrint(e)
			}
		}
	}
}
