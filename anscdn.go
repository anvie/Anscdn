/**
*
* 	AnsCDN Copyright (C) 2010 Robin Syihab (r@nosql.asia)
*	Simple CDN server written in Golang.
*
*	License: General Public License v2 (GPLv2)
*
**/


package main;


import (
	"io/ioutil"
	"strings"
	"strconv"
	"fmt"
	"http"
	"os"
	"path"
	"mime"
	"utf8"
	"flag"
	"./anlog"
	"./filemon"
	"./config"
	"./utils"
)

const (
	VERSION = "0.7 beta"
)

var cfg *config.AnscdnConf
var quiet bool

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

func setHeaderCond(con http.ResponseWriter, abs_path string, data []byte) {
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

func validUrlPath(url_path string) bool{
	return strings.Index(url_path,"../") < 1
}

func write(c http.ResponseWriter, f string, v ...interface{}){fmt.Fprintf(c,f,v...);}

func MainHandler(con http.ResponseWriter, r *http.Request){
	
	url_path := r.URL.Path[1:]
	
	if len(url_path) == 0{
		http.Error(con,"404",http.StatusNotFound)
		return
	}
	
	// security check
	if !validUrlPath(url_path){
		write(con,"Invalid url path")
		anlog.Warn("Invalid url_path: %s\n",url_path)
		return
	}
	
	// restrict no ext
	if cfg.IgnoreNoExt && len(path.Ext(url_path)) == 0 {
		anlog.Warn("Ignoring `%s`\n", url_path)
		http.Error(con, "404", http.StatusNotFound)
		return
	}
	
	// restrict ext
	if len(cfg.IgnoreExt) > 0 {
		cext := path.Ext(url_path)
		if len(cext) > 1{
			cext = strings.ToLower(cext[1:])
			exts := strings.Split(cfg.IgnoreExt,",",0)
			for _, ext := range exts{
				if cext == strings.Trim(ext," ") {
					anlog.Warn("Ignoring `%s` by extension.\n", url_path)
					http.Error(con, "404", http.StatusNotFound)
					return
				}
			}
		}
	}
	
	var abs_path string
	
	if strings.HasPrefix(cfg.StoreDir,"./"){
		
		abs_path, _ = os.Getwd()
		abs_path = path.Join(abs_path, cfg.StoreDir[1:], url_path)
		
	}else{
		abs_path = path.Join(cfg.StoreDir,url_path)
	}
	
	dir_name, _ := path.Split(abs_path)
	
	if !file_exists(abs_path) {
		
		url_source := "http://" + cfg.BaseServer + "/" + url_path
		
		err := os.MkdirAll(dir_name,0755)
		if err != nil {
			fmt.Fprintf(con,"404 Not found (e)")
			anlog.Error("Cannot MkdirAll. error: %s\n",err.String())
			return
		}
		
		// download it
		
		r, _, err := http.Get(url_source)
		if err != nil {
			anlog.Error("Cannot download data from `%s`. e: %s\n", url_source, err.String())
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
			ctype := utils.FixedMime(ctype)
			exttype := utils.FixedMime(ext_type)
			if exttype != ctype {
				anlog.Warn("Mime type different by extension. `%s` <> `%s` path `%s`\n", ctype, exttype, url_path )
				if cfg.Strict {
					http.Error(con, "404", http.StatusNotFound)
					return
				}
			}
		}
		
		anlog.Info("File `%s` first cached from `%s`.\n", abs_path, url_source)
		
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
		
		if cfg.CacheOnly {
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

func ClearCacheHandler(c http.ResponseWriter, r *http.Request){

	path_to_clear := r.FormValue("p")
	if len(path_to_clear) == 0{
		write(c,"Invalid parameter")
		return
	}

	// prevent canonical path
	if strings.HasPrefix(path_to_clear,"."){
		write(c,"Bad path")
		return
	}
	if path_to_clear[0] == '/'{
		path_to_clear = path_to_clear[1:]
	}
	path_to_clear = "./data/" + path_to_clear
	
	f, err := os.Open(path_to_clear,os.O_RDONLY,0)
	if err != nil{
		anlog.Error("File open error %s\n", err.String())
		write(c,"Invalid request")
		return
	}
	defer f.Close()
	
	st, err := f.Stat()
	if err != nil{
		anlog.Error("Cannot stat file. error %s\n", err.String())
		write(c,"Invalid request")
		return
	}
	if !st.IsDirectory(){
		write(c,"Invalid path")
		return
	}
	
	err = os.RemoveAll(path_to_clear)
	if err!=nil{
		write(c,"Cannot clear path. e: %s", err.String())
		return
	}
	
	store_dir := cfg.StoreDir
	
	if path_to_clear == store_dir + "/"{
		if err := os.Mkdir(store_dir,0775); err != nil{
			anlog.Error("Cannot recreate base store_dir: `%s`\n", store_dir)
		}
	}
	
	anlog.Info("Path cleared by request from `%s`: `%s`\n", r.Host, path_to_clear)
	write(c,"Clear successfully")
}

func intro(){
	fmt.Println("\n AnsCDN " + VERSION + " - a Simple CDN Server")
	fmt.Println(" Copyright (C) 2010 Robin Syihab (r@nosql.asia)")
	fmt.Println(" Under GPLv2 License\n")
}


func main() {
	
	intro()
	
	var cfg_file string
	
	flag.StringVar(&cfg_file,"config","anscdn.cfg","Config file.")
	flag.BoolVar(&quiet,"quiet",false,"Quiet.")
	
	flag.Parse()
	
	anlog.Quiet = quiet
	
	var err os.Error
	cfg, err = config.Parse(cfg_file)
	
	if err != nil {
		fmt.Println("Invalid configuration. e: ",err.String(),"\n")
		os.Exit(1)
	}
	
	if len(cfg.BaseServer) == 0{
		anlog.Error("No base server")
		os.Exit(3)
	}
	if cfg.ServingPort == 0{
		anlog.Error("No port")
		os.Exit(4)
	}
	if len(cfg.StoreDir) == 0{
		cfg.StoreDir = "./data"
	}
	if cfg.CacheExpires == 0{
		cfg.CacheExpires = 1296000
	}
	
	fmt.Println("Configuration:")
	fmt.Println("---------------------------------------")
	
	fmt.Println("Base server: " + cfg.BaseServer)
	if cfg.Strict == true {
		fmt.Println("Strict mode ON")
	}else{
		fmt.Println("Strict mode OFF")
	}
	if cfg.CacheOnly == true {
		fmt.Println("Cache only")
	}
	if cfg.IgnoreNoExt==true{fmt.Println("Ignore no extension files");}
	if len(cfg.IgnoreExt)>0{fmt.Println("Ignore extension for", cfg.IgnoreExt);}
	if len(cfg.ClearCachePath) > 0{
		fmt.Println("Clear cache path: ", cfg.ClearCachePath)
	}
	
	fmt.Printf("Store cached data in `%s`\n", cfg.StoreDir)
	
	if cfg.FileMon == true {
		fmt.Println("File monitor enabled")
		if err != nil{
			anlog.Error("Invalid cache_expires value `%d`\n", cfg.CacheExpires)
			os.Exit(5)
		}
		go filemon.StartFileMon(cfg.StoreDir, cfg.CacheExpires)
	}
	
	fmt.Println("---------------------------------------\n")
	
	anlog.Info("Serving on 0.0.0.0:%d... ready.\n", cfg.ServingPort )
	
	
	if len(cfg.ClearCachePath) > 0 {
		if cfg.ClearCachePath[0] != '/'{
			anlog.Error("Invalid ccp `%s`. missing `/`\n",cfg.ClearCachePath)
			os.Exit(2)
		}
		http.Handle(cfg.ClearCachePath, http.HandlerFunc(ClearCacheHandler))
	}
	http.Handle("/", http.HandlerFunc(MainHandler))
	if err := http.ListenAndServe("0.0.0.0:" + strconv.Itoa(cfg.ServingPort), nil); err != nil {
		anlog.Error("%s\n",err.String())
	}
}

