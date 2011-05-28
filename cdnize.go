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
	"strings"
	"json"
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

func Jsonize(data interface{}) string{
	rv, err := json.Marshal(&data)
	if err != nil{
		anlog.Error("Cannot jsonize `%v`\n", data)
		return ""
	}
	return string(rv)
}

func jsonError(status, info string) string{
	return Jsonize(map[string]string{"Status":status,"Info":info})
}

func Handler(c http.ResponseWriter, r *http.Request){

	c.Header().Set("Content-Type", "application/json")
	
	api_key := r.FormValue("api_key")
	//base_url := r.FormValue("base_url")
	
	if api_key != Cfg.ApiKey{
		write(c,jsonError("failed","Invalid api key"))
		return
	}
	
	requested_url := r.FormValue("u")
	if requested_url == ""{
		
		// no url
		
		r.ParseForm()
		file_name := r.FormValue("file_name")

		if file_name == ""{
			write(c,jsonError("failed","no `file_name` parameter"))
			return
		}
		
		fmt.Printf("file_name: %v\n", file_name)
		
		file, err := r.MultipartReader()
		if err != nil{
			write(c,jsonError("failed","cannot get multipart reader"))
			return	
		}

		part, err := file.NextPart()
		if err != nil{
			write(c,jsonError("failed","no `u` nor `file`"))
			return
		}
		var data [1000]byte
		var i int = 0
		var data_size int64 = 0
		md5ed := hmac.NewMD5([]byte("cdnized-2194"))
		abs_path := "/tmp/" + RandStrings(100)
		dst_file, err := os.OpenFile(abs_path,os.O_WRONLY | os.O_CREATE,0755)
		if err != nil {
			anlog.Error("Cannot create file `%s`. error: %s\n", abs_path,err.String())
			write(c,jsonError("failed",fmt.Sprintf("cannot create temporary data. %v\n",err)))
			return
		}

		for data_size < r.ContentLength{
			i, err = part.Read(data[0:999])
			if err !=nil{
				break
			}
			
			_, err := md5ed.Write(data[0:i])
			if err != nil{
				anlog.Error("Cannot calculate MD5 hash")
				write(c,jsonError("failed","cannot calculate checksum"))
				break
			}
			
			_, err = dst_file.Write(data[0:i])
			if err != nil{
				anlog.Error("Cannot write %d bytes data in file `%s`. error: %s\n", data_size, abs_path, err.String())
			}
			
			data_size += int64(i)
		}
		
		dst_file.Close()
		
		//fmt.Printf("content-length: %v, file: %v, file-length: %v, i: %v\n", r.ContentLength, string(data[0:]), i, i)
		
		hash := fmt.Sprintf("%x", md5ed.Sum())
		file_ext := strings.ToLower(path.Ext(file_name))
		file_name = hash + file_ext
		new_path, err := os.Getwd()
		
		new_path = path.Join(new_path, Cfg.StoreDir[2:], Cfg.ApiStorePrefix, file_name)
		
		if err != nil {
			anlog.Error("Cannot getwd\n")
			write(c,jsonError("failed","internal error"))
			return
		}
		
		//fmt.Printf("abs_path: %v, new_path: %v\n", abs_path, new_path)
		if err := syscall.Rename(abs_path, new_path); err != 0{
			anlog.Error("Cannot move from file `%s` to `%s`. %v.\n", abs_path, new_path, err)
			write(c,jsonError("failed","internal error"))
			return
		}
		
		cdnized_url := fmt.Sprintf("http://%s/%s/%s/%s", Cfg.CdnServerName, Cfg.StoreDir[2:], Cfg.ApiStorePrefix, file_name)

		anlog.Info("cdnized_url: %s\n", cdnized_url)
		
		os.Remove(abs_path)
		
		
		type success struct{
			Status string
			Size int64
			Cdnized_url string
		}

		write(c, Jsonize(&success{"ok", data_size, cdnized_url}))
		return
	}
	
	// with url
	
	//write(c, fmt.Sprintf("{Status: 'ok', url_path: '%s', gen: '%s'}", requested_url, x))
	
	file_ext := strings.ToLower(path.Ext(requested_url))
	abs_path, _ := os.Getwd()
	abs_path = path.Join(abs_path, Cfg.StoreDir[2:], Cfg.ApiStorePrefix, RandStrings(64) + file_ext)
	
	fmt.Printf("abs_path: %s\n", abs_path)
	
	var data []byte;
	rv, lm, tsize := downloader.Download(requested_url, abs_path, true, &data)
	if rv != true{
		write(c,jsonError("failed","Cannot fetch from source url"))
		return
	}
	
	md5ed := hmac.NewMD5([]byte("cdnized-2194"))
	for {
		brw, err := md5ed.Write(data)
		if err != nil{
			anlog.Error("Cannot calculate MD5 hash")
			write(c,jsonError("failed","Internal error"))
			return
		}
		if brw >= tsize{
			break;
		}
	}
	
	hash := fmt.Sprintf("%x", md5ed.Sum())
	dir, _ := path.Split(abs_path)
	file_name := hash + file_ext
	new_path := path.Join(dir, file_name)
	
	if err := syscall.Rename(abs_path, new_path); err != 0{
		anlog.Error("Cannot rename from file `%s` to `%s`", abs_path, new_path)
		write(c,jsonError("failed","Internal error"))
		return
	}
	
	cdnized_url := fmt.Sprintf("http://%s/%s/%s/%s", Cfg.CdnServerName, Cfg.StoreDir[2:], Cfg.ApiStorePrefix, file_name)
	
	anlog.Info("cdnized_url: %s", cdnized_url)
	
	type success struct{
		Status string
		Lm string
		Size int
		Original string
		Cdnized_url string
	}
	
	write(c, Jsonize(&success{"ok", lm, tsize, requested_url, cdnized_url}))
}

func StaticHandler(c http.ResponseWriter, r *http.Request){
	path := r.URL.Path
	root, _ := os.Getwd()
	http.ServeFile(c, r, root + "/" + path)
}


