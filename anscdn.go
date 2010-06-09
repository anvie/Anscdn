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

func MainHandler(con *http.Conn, r *http.Request){
	
	url_path := r.URL.Path[1:]
	
	abs_path, _ := os.Getwd()
	abs_path += "/" + url_path
	
	dir_name, _ := path.Split(abs_path)
	
	if !file_exists(abs_path) {
		
		url_source := base_server + "/" + url_path
		
		fmt.Printf("File `%s` first cached from `%s`.\n", abs_path, url_source)
		
		err := os.MkdirAll(dir_name,0755)
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			fmt.Printf("Cannot MkdirAll. error: %s\n",err.String())
			return
		}
		
		// download it
		
		r, _, err := http.Get(url_source)
		if err != nil {
			fmt.Printf("[error] Cannot download data form `%s`\n", url_source)
			fmt.Fprintf(con,"404 Not found (e)")
			return
		}
		
		var data []byte
		data, err = ioutil.ReadAll(r.Body)
		
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			fmt.Printf("Cannot read url source body `%s`. error: %s\n", abs_path,err.String())
			return
		}
		
		file, err := os.Open(abs_path,os.O_WRONLY | os.O_CREAT,0755)
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			fmt.Printf("Cannot create file `%s`. error: %s\n", abs_path,err.String())
			return
		}
		defer file.Close()
		
		var total_size int = len(data)
		for {
			bw, err := file.Write(data)
			if err != nil {
				fmt.Fprintf(con,"404 Not found (e)")
				fmt.Printf("Cannot write %d bytes data in file `%s`. error: %s\n", total_size, abs_path,err.String())
				return
			}
			if bw >= total_size {
				break
			}
		}
		
		// send to client for the first time.
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
		
		for {
			bw, err := con.Write(data)
			if err != nil {
				fmt.Fprintf(con,"404 Not found (e)")
				fmt.Printf("Cannot write %d bytes data in connection stream. error: %s\n", total_size,err.String())
				return
			}
			if bw >= total_size {
				break
			}
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
	
	flag.Parse()
	
	if bs == "" {
		fmt.Println("Invalid parameter.\n")
		os.Exit(1)
	}
	
	base_server = "http://" + bs
	
	fmt.Println("Serving on 0.0.0.0:" + serving_port )
	fmt.Println("Base server: " + bs)
	
	http.Handle("/", http.HandlerFunc(MainHandler))
	http.ListenAndServe(":" + serving_port, nil)
}

