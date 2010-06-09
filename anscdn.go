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
)

const (
	VERSION = "0.1 alpha"
)

var base_server string
var serving_port string
var strict bool
var cache_only bool

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

func Info(format string, v ...interface{}){fmt.Printf("[info] " + format, v);}
func Warn(format string, v ...interface{}) {fmt.Printf("[warning] " + format, v);}
func Error(format string, v ...interface{}) {fmt.Fprintf(os.Stderr,"[error] " + format, v);}

func MainHandler(con *http.Conn, r *http.Request){
	
	url_path := r.URL.Path[1:]
	
	abs_path, _ := os.Getwd()
	abs_path += "/" + url_path
	
	dir_name, _ := path.Split(abs_path)
	
	if !file_exists(abs_path) {
		
		url_source := base_server + "/" + url_path
		
		Info("File `%s` first cached from `%s`.\n", abs_path, url_source)
		
		err := os.MkdirAll(dir_name,0755)
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			Error("Cannot MkdirAll. error: %s\n",err.String())
			return
		}
		
		// download it
		
		r, _, err := http.Get(url_source)
		if err != nil {
			Error("Cannot download data form `%s`\n", url_source)
			fmt.Fprintf(con,"404 Not found (e)")
			return
		}
		
		var data []byte
		data, err = ioutil.ReadAll(r.Body)
		
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			Error("Cannot read url source body `%s`. error: %s\n", abs_path,err.String())
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
			if ext_type != ctype {
				Warn("Mime type different by extension. `%s` <> `%s` path `%s`\n", ctype, ext_type, url_path )
				if strict {
					http.Error(con, "404", http.StatusNotFound)
					return
				}
			}
		}
		
		file, err := os.Open(abs_path,os.O_WRONLY | os.O_CREAT,0755)
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			Error("Cannot create file `%s`. error: %s\n", abs_path,err.String())
			return
		}
		defer file.Close()
		
		var total_size int = len(data)
		for {
			bw, err := file.Write(data)
			if err != nil {
				fmt.Fprintf(con,"404 Not found (e)")
				Error("Cannot write %d bytes data in file `%s`. error: %s\n", total_size, abs_path,err.String())
				return
			}
			if bw >= total_size {
				break
			}
		}
		
		// send to client for the first time.
		setHeaderCond(con, abs_path, data)

		for {
			bw, err := con.Write(data)
			if err != nil {
				fmt.Fprintf(con,"404 Not found (e)")
				Error("Cannot send file `%s`. error: %s\n", abs_path, err.String())
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
			Error("Cannot open file `%s`. error: %s\n", abs_path,err.String())
			return
		}
		
		defer file.Close()
		
		bufsize := 1024
		buff := make([]byte,bufsize+2)
		
		_, err = file.Read(buff)
		
		if err != nil && err != os.EOF {
			fmt.Fprintf(con,"404 Not found (e)")
			Error("Cannot read %d bytes data in file `%s`. error: %s\n", bufsize, abs_path,err.String())
			return 
		}
		
		setHeaderCond(con, abs_path, buff)
		
		con.Write(buff)
		
		for {
			
			_, err := file.Read(buff)
			if err !=nil {
				if err == os.EOF {
					break
				}
				fmt.Fprintf(con,"404 Not found (e)")
				Error("Cannot read %d bytes data in file `%s`. error: %s\n", bufsize, abs_path,err.String())
				return
			}
			con.Write(buff)
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
	flag.BoolVar(&strict,"strict",true,"Strict mode. Don't cache invalid mime type.")
	flag.BoolVar(&cache_only,"cache-only",false,"Don't serve cached file.")
	
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
	
	Info("Serving on 0.0.0.0:" + serving_port + "... ready.\n" )
	
	http.Handle("/", http.HandlerFunc(MainHandler))
	http.ListenAndServe(":" + serving_port, nil)
}

