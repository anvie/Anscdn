

package cdnize

import (
	"fmt"
	"http"
	"rand"
	"time"
	"path"
	"crypto/hmac"
	"os"
	"syscall"
	"./anlog"
	"./config"
	"./downloader"
)

var Cfg *config.AnscdnConf

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123567890abcdefghijklmnopqrstuvwxyz_";

func RandStrings(N int) string {
	rand.Seed(time.Nanoseconds())

	buf := make([]byte, N + 1)

	for i := 0; i < N; i++ {
		buf[i] = chars[rand.Intn(len(chars))]
	}
	return string(buf[0:N])
}

func write(c http.ResponseWriter, f string, v ...interface{}){fmt.Fprintf(c,f,v...);}


func Handler(c http.ResponseWriter, r *http.Request){

	api_key := r.FormValue("api_key")
	
	if api_key != Cfg.ApiKey{
		write(c,"{'status': 'invalid Api key'}")
		return
	}
	
	requested_url := r.FormValue("u")
	if requested_url == ""{
		
		r.ParseForm()
		file_name := r.FormValue("file_name")

		if file_name == ""{
			write(c,"{'status': 'failed','info': 'no `file_name` parameter'}")
			return
		}
		
		fmt.Printf("file_name: %v\n", file_name)
		
		file, err := r.MultipartReader()
		if err != nil{
			write(c,"{'status': 'failed','info': 'cannot get multipart reader'}")
			return	
		}

		part, err := file.NextPart()
		if err != nil{
			write(c,"{'status': 'failed, no `u` nor `file`'}")
			return
		}
		var data [1000]byte
		var i int = 0
		var data_size int64 = 0
		md5ed := hmac.NewMD5([]byte("cdnized-2194"))
		abs_path := "/tmp/" + RandStrings(100)
		dst_file, err := os.Open(abs_path,os.O_WRONLY | os.O_CREAT,0755)
		if err != nil {
			anlog.Error("Cannot create file `%s`. error: %s\n", abs_path,err.String())
			write(c,"{'status': 'failed, cannot create temporary data'}")
			return
		}

		for data_size < r.ContentLength{
			i, err = part.Read(data[0:999])
			if err !=nil{
				anlog.Error("Cannot read more part. %s.\n", err)
				break
			}
			
			_, err := md5ed.Write(data[0:i])
			if err != nil{
				anlog.Error("Cannot calculate MD5 hash")
				write(c,"{status: 'failed'}")
				break
			}
			
			_, err = dst_file.Write(data[0:i])
			if err != nil{
				anlog.Error("Cannot write %d bytes data in file `%s`. error: %s\n", data_size, abs_path, err.String())
			}
			
			data_size += int64(i)
		}
		
		dst_file.Close()
		
		fmt.Printf("content-length: %v, file: %v, file-length: %v, i: %v\n", r.ContentLength, string(data[0:]), i, i)
		
		hash := fmt.Sprintf("%x", md5ed.Sum())
		file_ext := path.Ext(file_name)
		file_name = hash + "_2194_" + RandStrings(100) + file_ext
		new_path, err := os.Getwd()
		
		new_path = path.Join(new_path, Cfg.StoreDir[2:], file_name)
		
		if err != nil {
			anlog.Error("Cannot getwd\n")
			write(c, "{'status': 'failed'}")
			return
		}
		
		//fmt.Printf("abs_path: %v, new_path: %v\n", abs_path, new_path)
		if err := syscall.Rename(abs_path, new_path); err != 0{
			anlog.Error("Cannot move from file `%s` to `%s`. %v.\n", abs_path, new_path, err)
			write(c,"{status: 'failed'}")
			return
		}
		
		cdnized_url := fmt.Sprintf("http://%s/%s/%s", Cfg.CdnServerName, Cfg.StoreDir[2:], file_name)

		anlog.Info("cdnized_url: %s", cdnized_url)
		
		os.Remove(abs_path)

		write(c, fmt.Sprintf("{'status': 'ok', 'size': '%v', 'original': '%s', 'cdnized_url': '%s'}", data_size, requested_url, cdnized_url))
		return
	}
	
	
	
	//write(c, fmt.Sprintf("{status: 'ok', url_path: '%s', gen: '%s'}", requested_url, x))
	
	file_ext := path.Ext(requested_url)
	abs_path, _ := os.Getwd()
	abs_path = path.Join(abs_path, Cfg.StoreDir[2:], RandStrings(64) + file_ext)
	
	fmt.Printf("abs_path: %s\n", abs_path)
	
	var data []byte;
	rv, lm, tsize := downloader.Download(requested_url, abs_path, true, &data)
	if rv != true{
		write(c,"{status: 'failed'}")
		return
	}
	
	md5ed := hmac.NewMD5([]byte("cdnized-2194"))
	for {
		brw, err := md5ed.Write(data)
		if err != nil{
			anlog.Error("Cannot calculate MD5 hash")
			write(c,"{status: 'failed'}")
			return
		}
		if brw >= tsize{
			break;
		}
	}
	
	hash := fmt.Sprintf("%x", md5ed.Sum())
	dir, _ := path.Split(abs_path)
	file_name := hash + "_2194_" + RandStrings(8) + file_ext
	new_path := path.Join(dir, file_name)
	
	if err := syscall.Rename(abs_path, new_path); err != 0{
		anlog.Error("Cannot rename from file `%s` to `%s`", abs_path, new_path)
		write(c,"{status: 'failed'}")
		return
	}
	
	cdnized_url := fmt.Sprintf("http://%s/%s/%s", Cfg.CdnServerName, Cfg.StoreDir[2:], file_name)
	
	anlog.Info("cdnized_url: %s", cdnized_url)
	
	write(c, fmt.Sprintf("{status: 'ok', lm: '%s', size: '%v', original: '%s', cdnized_url: '%s'}", lm, tsize, requested_url, cdnized_url))
}

func StaticHandler(c http.ResponseWriter, r *http.Request){
	path := r.URL.Path
	root, _ := os.Getwd()
	http.ServeFile(c, r, root + "/" + path)
}


