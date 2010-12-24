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

package anlog



import (
	"fmt"
	"os"
)

var Quiet bool

func Info(format string, v ...interface{}) {
	if Quiet{return;}
	fmt.Printf("[info] " + format, v...);
}
func Warn(format string, v ...interface{}) {fmt.Printf("[warning] " + format, v...);}
func Error(format string, v ...interface{}) {fmt.Fprintf(os.Stderr,"[error] " + format, v...);}

