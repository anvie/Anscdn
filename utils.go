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

package utils


var FixedMimeList = map[string]string{
	"js" : "application/x-javascript",
}

var VariantMimeList = map[string]string{
	"application/javascript" : FixedMimeList["js"],
}


func FixedMime(mimeType string) string {
	if v, ok := VariantMimeList[mimeType]; ok{
		return v
	}
	return mimeType
}
