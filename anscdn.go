/**
*
* 	AnsCDN Copyright (C) 2010 Robin Syihab (r@nosql.asia)
*	Simple CDN server written in Golang.
*
*	License: General Public License v2 (GPLv2)
*	version 0.1 alpha
*
**/


package main;


import (
	"io/ioutil"
	"strings"
	"fmt"
	"http"
	"os"
	"path"
	"mime"
	"utf8"
	"flag"
	"./anlog"
	"./filemon"
)

const (
	VERSION = "0.2 beta"
)

var base_server string
var serving_port string
var store_dir string
var strict bool
var cache_only bool
var file_mon bool
var cache_expires int64

func file_exists(file_path string) bool{
	file, err := os.Open(file_path, os.O_RDONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// Heuristic: b is text if it is valid UTF-8 and doesn't
// contain any unprintable ASCII or Unicode characters.
func isText(b []byte) bool {
    for len(b) > 0 && utf8.FullRune(b) {
        rune, size := utf8.DecodeRune(b)
        if size == 1 && rune == utf8.RuneError {
            // decoding error
            return false
        }
        if 0x80 <= rune && rune <= 0x9F {
            return false
        }
        if rune < ' ' {
            switch rune {
            case '\n', '\r', '\t':
                // okay
            default:
                // binary garbage
                return false
            }
        }
        b = b[size:]
    }
    return true
}

func setHeaderCond(con *http.Conn, abs_path string, data []byte) {
	extension := path.Ext(abs_path)
	if ctype := mime.TypeByExtension(extension); ctype != "" {
		con.SetHeader("Content-Type", ctype)
	}else{
        if isText(data) {
            con.SetHeader("Content-Type", "text-plain; charset=utf-8")
        } else {
            con.SetHeader("Content-Type", "application/octet-stream") // generic binary
        }
	}
	
}


func MainHandler(con *http.Conn, r *http.Request){
	
	url_path := r.URL.Path[1:]
	
	var abs_path string
	
	if strings.HasPrefix(store_dir,"./"){
		
		abs_path, _ = os.Getwd()
		abs_path = path.Join(abs_path, store_dir[1:], url_path)
		
	}else{
		abs_path = path.Join(store_dir,url_path)
	}
	
	dir_name, _ := path.Split(abs_path)
	
	if !file_exists(abs_path) {
		
		url_source := base_server + "/" + url_path
		
		anlog.Info("File `%s` first cached from `%s`.\n", abs_path, url_source)
		
		err := os.MkdirAll(dir_name,0755)
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			anlog.Error("Cannot MkdirAll. error: %s\n",err.String())
			return
		}
		
		// download it
		
		r, _, err := http.Get(url_source)
		if err != nil {
			anlog.Error("Cannot download data form `%s`\n", url_source)
			fmt.Fprintf(con,"404 Not found (e)")
			return
		}
		
		var data []byte
		data, err = ioutil.ReadAll(r.Body)
		
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			anlog.Error("Cannot read url source body `%s`. error: %s\n", abs_path,err.String())
			return
		}
		
		// check for the mime
		ctype := r.Header["Content-Type"]
		if endi := strings.IndexAny(ctype,";"); endi > 1 {
			ctype = ctype[0:endi]
		}else{
			ctype = ctype[0:]
		}
		
		// fmt.Printf("Content-type: %s\n",ctype)
		if ext_type := mime.TypeByExtension(path.Ext(abs_path)); ext_type != "" {
			
			if endi := strings.IndexAny(ext_type,";"); endi > 1 {
				ext_type = ext_type[0:endi]
			}else{
				ext_type = ext_type[0:]
			}
			
			if ext_type != ctype {
				anlog.Warn("Mime type different by extension. `%s` <> `%s` path `%s`\n", ctype, ext_type, url_path )
				if strict {
					http.Error(con, "404", http.StatusNotFound)
					return
				}
			}
		}
		
		file, err := os.Open(abs_path,os.O_WRONLY | os.O_CREAT,0755)
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			anlog.Error("Cannot create file `%s`. error: %s\n", abs_path,err.String())
			return
		}
		defer file.Close()
		
		var total_size int = len(data)
		for {
			bw, err := file.Write(data)
			if err != nil {
				fmt.Fprintf(con,"404 Not found (e)")
				anlog.Error("Cannot write %d bytes data in file `%s`. error: %s\n", total_size, abs_path,err.String())
				//file.Close()
				return
			}
			if bw >= total_size {
				break
			}
		}
		
		//file.Close()
		
		// send to client for the first time.
		setHeaderCond(con, abs_path, data)
		
		// set Last-modified header
		lm,_ := filemon.GetLastModif(file)
		con.SetHeader("Last-Modified", lm)
		
		for {
			bw, err := con.Write(data)
			if err != nil {
				fmt.Fprintf(con,"404 Not found (e)")
				anlog.Error("Cannot send file `%s`. error: %s\n", abs_path, err.String())
				return
			}
			if bw >= total_size {
				break
			}
		}
		
	}else{
		
		if cache_only {
			// no static serving, use external server like nginx etc.
			return
		}
		
		// if file exists, just send it
		
		file, err := os.Open(abs_path,os.O_RDONLY,0)
		
		if err != nil{
			fmt.Fprintf(con,"404 Not found (e)")
			anlog.Error("Cannot open file `%s`. error: %s\n", abs_path,err.String())
			return
		}
		
		defer file.Close()
		
		bufsize := 1024*4
		buff := make([]byte,bufsize+2)

		sz, err := file.Read(buff)
		
		if err != nil && err != os.EOF {
			fmt.Fprintf(con,"404 Not found (e)")
			anlog.Error("Cannot read %d bytes data in file `%s`. error: %s\n", sz, abs_path,err.String())
			return 
		}
		
		setHeaderCond(con, abs_path, buff)
		
		// check for last-modified
		//r.Header["If-Modified-Since"]
		lm, _ := filemon.GetLastModif(file)
		con.SetHeader("Last-Modified", lm)
		
		if r.Header["If-Modified-Since"] == lm {
			con.WriteHeader(http.StatusNotModified)
			return
		}
		
		con.Write(buff[0:sz])
		
		for {
			sz, err := file.Read(buff)
			if err != nil {
				if err == os.EOF {
					con.Write(buff[0:sz])
					break
				}
				fmt.Fprintf(con,"404 Not found (e)")
				anlog.Error("Cannot read %d bytes data in file `%s`. error: %s\n", sz, abs_path,err.String())
				return
			}
			con.Write(buff[0:sz])
		}
		
	}
	
}

func intro(){
	fmt.Println("\n AnsCDN " + VERSION + " - a Simple CDN Server")
	fmt.Println(" Copyright (C) 2010 Robin Syihab (r@nosql.asia)")
	fmt.Println(" Under GPLv2 License\n")
}


func main() {
	
	intro()
	
	var bs string
	
	flag.StringVar(&bs,"base-server","","Base server address ex: 127.0.0.1:2194.")
	flag.StringVar(&serving_port,"port","2194","Serving port.")
	flag.StringVar(&store_dir,"store-dir","./data","Location path to store cached data.")
	flag.BoolVar(&strict,"strict",true,"Strict mode. Don't cache invalid mime type.")
	flag.BoolVar(&cache_only,"cache-only",false,"Don't serve cached file.")
	flag.BoolVar(&file_mon,"file-mon",true,"Enable file cache monitor.")
	flag.Int64Var(&cache_expires,"cx",1296000,"File deleted automatically after seconds time not accessed.\nWork only if `file-mon` enabled")
	
	flag.Parse()
	
	if bs == "" {
		fmt.Println("Invalid parameter.\n")
		os.Exit(1)
	}
	
	base_server = "http://" + bs
	
	fmt.Println("Base server: " + bs)
	if strict {
		fmt.Println("Strict mode ON")
	}else{
		fmt.Println("Strict mode OFF")
	}
	if cache_only{
		fmt.Println("Cache only")
	}
	if file_mon{
		fmt.Println("File monitor enabled")
		go filemon.StartFileMon(store_dir, cache_expires)
	}
	
	fmt.Printf("Store cached data in `%s`\n", store_dir)
	
	anlog.Info("Serving on 0.0.0.0:" + serving_port + "... ready.\n" )
	
	http.Handle("/", http.HandlerFunc(MainHandler))
	http.ListenAndServe("0.0.0.0:" + serving_port, nil)
}

