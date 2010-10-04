/**
*
* 	AnsCDN Copyright (C) 2010 Robin Syihab (r@nosql.asia)
*	Simple CDN server written in Golang.
*
*	License: General Public License v2 (GPLv2)
*
**/

package config


import (
	"os"
	"./configfile"
)

type AnscdnConf struct {
	BaseServer string
	ServingPort int
	StoreDir string
	Strict bool
	CacheOnly bool
	FileMon bool
	CacheExpires int64
	ClearCachePath string
	IgnoreNoExt bool
	IgnoreExt string	
}

func Parse(file string) (ac *AnscdnConf, err os.Error) {

	conf, err := configfile.ReadConfigFile(file)
	
	if err != nil{
		return nil, err
	}
	
	BaseServer, err := conf.GetString("default","base_server")
	ServingPort, err := conf.GetInt("default","serving_port")
	StoreDir, err := conf.GetString("default","store_dir")
	Strict, err := conf.GetBool("default","strict")
	CacheOnly, err := conf.GetBool("default","cache_only")
	FileMon, err := conf.GetBool("default","file_mon")
	CacheExpires, err := conf.GetInt64("default","cache_expires")
	ClearCachePath, err := conf.GetString("default","clear_cache_path")
	IgnoreNoExt, err := conf.GetBool("default","ignore_no_ext")
	IgnoreExt, err := conf.GetString("default","ignore_ext")

	return &AnscdnConf{BaseServer,
		ServingPort,StoreDir,Strict,CacheOnly,FileMon,CacheExpires,ClearCachePath,IgnoreNoExt,IgnoreExt}, err
}
