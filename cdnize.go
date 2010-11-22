

package cdnize

import (
	"fmt"
	"http"
	"rand"
	"time"
	//"./downloader"
)

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
	return string(buf[0:N])
}

func write(c http.ResponseWriter, f string, v ...interface{}){fmt.Fprintf(c,f,v...);}

func Handler(c http.ResponseWriter, r *http.Request){
	x := RandStrings(64)
	write(c, fmt.Sprintf("{status: 'ok', url_path: '%s', gen: '%s'}", r.FormValue("u"), x))
}

