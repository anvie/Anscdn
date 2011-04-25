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

package downloader

import (
	"io/ioutil"
	"strings"
	"http"
	"os"
	"path"
	"mime"
	"./filemon"
	"./anlog"
	"./utils"
)

func Download(url_source string, abs_path string, strict bool, data *[]byte) (rv bool, lm string, total_size int) {

	resp, _, err := http.Get(url_source)
	if err != nil {
		anlog.Error("Cannot download data from `%s`. e: %s\n", url_source, err.String())
		return false, "", 0
	}

	*data, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		anlog.Error("Cannot read url source body `%s`. error: %s\n", abs_path,err.String())
		return false, "", 0
	}

	// check for the mime
	content_type := resp.Header.Get("Content-Type")
	if endi := strings.IndexAny(content_type,";"); endi > 1 {
		content_type = content_type[0:endi]
	}else{
		content_type = content_type[0:]
	}

	// fmt.Printf("Content-type: %s\n",ctype)
	if ext_type := mime.TypeByExtension(path.Ext(abs_path)); ext_type != "" {
		if endi := strings.IndexAny(ext_type,";"); endi > 1 {
			ext_type = ext_type[0:endi]
		}else{
			ext_type = ext_type[0:]
		}
		content_type := utils.FixedMime(content_type)
		exttype := utils.FixedMime(ext_type)
		if exttype != content_type {
			anlog.Warn("Mime type different by extension. `%s` <> `%s` path `%s`\n", content_type, exttype, url_source )
			if strict {
				return false, "", 0
			}
		}
	}

	anlog.Info("File `%s` first cached from `%s`.\n", abs_path, url_source)

	file, err := os.OpenFile(abs_path,os.O_WRONLY | os.O_CREATE,0755)
	if err != nil {
		anlog.Error("Cannot create file `%s`. error: %s\n", abs_path,err.String())
		return false, "", 0
	}
	defer file.Close()

	total_size = len(*data)
	for {
		bw, err := file.Write(*data)
		if err != nil {
			anlog.Error("Cannot write %d bytes data in file `%s`. error: %s\n", total_size, abs_path,err.String())
			return false, "", 0
		}
		if bw >= total_size {
			break
		}
	}
	
	lm, _ = filemon.GetLastModif(file)
	
	return true, lm, total_size
}
