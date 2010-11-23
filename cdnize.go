

package cdnize

import (
	"fmt"
	"http"
	"rand"
	"time"
	"path"
	"crypto/md5"
	"os"
	"./config"
	"./downloader"
)

var Cfg *config.AnscdnConf

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz_";

func RandStrings(N int) string {
	rand.Seed(time.Nanoseconds())
	//r := make([]string, N)
	//ri := 0
	buf := make([]byte, N + 1)
	//known := map[string]bool{}

	for i := 0; i < N; i++ {
		buf[i] = chars[rand.Intn(len(chars))]
	}
	rv := string(buf[0:N])
	
	hasher := md5.New()
	if _, err := hasher.Write([]byte(fmt.Sprintf("%x",time.Nanoseconds()))); err != nil{
		return rv;
	}
	hash := string(hasher.Sum())
	
	return rv + "_" + fmt.Sprintf("%x",hash)
}

func write(c http.ResponseWriter, f string, v ...interface{}){fmt.Fprintf(c,f,v...);}

func Handler(c http.ResponseWriter, r *http.Request){

	requested_url := r.FormValue("u")
	if requested_url == ""{
		write(c,"{status: 'failed'}")
		return
	}
	//write(c, fmt.Sprintf("{status: 'ok', url_path: '%s', gen: '%s'}", requested_url, x))
	
	abs_path, _ := os.Getwd()
	cdnized_url := fmt.Sprintf("/%s/%s%s", Cfg.StoreDir, RandStrings(64), path.Ext(requested_url))
	abs_path = path.Join(abs_path, cdnized_url)
	
	fmt.Printf("abs_path: %s\n", abs_path)
	fmt.Printf("cdnized_url: %s\n", cdnized_url)
	
	var data []byte;
	rv, lm, tsize := downloader.Download(requested_url, abs_path, true, &data)
	if rv != true{
		write(c,"{status: 'failed'}")
		return
	} 
	
	write(c, fmt.Sprintf("{status: 'ok', lm: '%s', size: '%v', original: '%s', cdnized_url: '%s'}", lm, tsize, requested_url, cdnized_url))
}

func StaticHandler(c http.ResponseWriter, r *http.Request){
	path := r.URL.Path
	root, _ := os.Getwd()
	http.ServeFile(c, r, root + "/" + path)
}


