/**
*
* 	AnsCDN Copyright (C) 2010 Robin Syihab (r [at] nosql.asia)
*	Simple CDN server written in Golang.
*
*	License: General Public License v2 (GPLv2)
*
*	Copyright (c) 2009 The Go Authors. All rights reserved.
*
**/

package filemon


import (
	"os"
	"time"
	//"fmt"
	"path"
	"strings"
	"./anlog"
)


var cache_expires int64


func isObsolete(atime_ns int64) (rv bool, old int64) {
	//tf := time.SecondsToLocalTime(atime_ns / 1e9)
	//tn := time.SecondsToLocalTime(time.Seconds() - 1296000)
	//anlog.Info("ft = %s\n", tf.String())
	//anlog.Info("nt = %s\n", tn.String())
	tndelta := (time.Seconds() - cache_expires)
	old = (tndelta - (atime_ns / 1e9))
	return (atime_ns < (tndelta * 1e9)), old
}


func rmObsolete(fpath string){
	
	f, err := os.Open(fpath)
	if err != nil{return;}
	
	defer f.Close()
	
	st, err := f.Stat()
	
	if err != nil {return;}
	
	if r, old := isObsolete(st.Atime_ns); r == true {
		
		anlog.Info("File `%s` is obsolete, %d seconds old.\n", fpath, old)
		anlog.Info("Delete file `%s`\n", fpath)
		
		if err := os.Remove(fpath); err != nil {
			anlog.Error("Cannot delete file `%s`. e: %s\n", err.String())
		}
		
	}

}


func processDir(p string){
	

	dir, err := os.Open(p)

	if err != nil {
	   anlog.Error("Cannot read dir `%s`. e: %s\n", p, err.String())
	   return
	}
	
	defer dir.Close()
	
	files, err := dir.Readdirnames(10)

	if err != nil {
	   anlog.Error("Cannot read dir `%s`. e: %s\n", p, err.String())
	   return
	}
	
	for _, f := range files{
		if strings.HasPrefix(f,".DS_"){
			continue
		}
		pp := path.Join(p,f)
		//fmt.Println(pp)
		o, err := os.Open(pp)
		if err != nil {
			continue
		}
		defer o.Close()
		if st,_:=o.Stat(); st != nil{
			if st.IsDirectory(){
				processDir(pp)
			}else{
				go rmObsolete(pp)
			}
		}
	}
	
}

func StartFileMon(store_dir string, cx int64){
	
	cache_expires = cx
	
	anlog.Info("File monitor started. (`%s`)\n", store_dir)

	for {
		
		//anlog.Info("Starting file auto cleaner...\n")
		
		processDir(store_dir)
		
		time.Sleep(86400 * 1e9) // 1x per day
	}
	
}

func GetLastModif(file *os.File) (rv string, err os.Error) {
	rv = ""
	st, err := file.Stat()
	if err != nil {return rv, err;}

	t := time.SecondsToLocalTime( st.Mtime_ns / 1e9 )
 	rv = t.Format(time.RFC1123)
	return rv, err
}

/*
func main(){
	
	processDir("./data")
	
}
*/





