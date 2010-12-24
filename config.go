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

package config


import (
	"os"
	"./configfile"
)

type AnscdnConf struct {
	BaseServer string
	ServingPort int
	StoreDir string
	ApiStorePrefix string
	Strict bool
	CacheOnly bool
	FileMon bool
	CacheExpires int64
	ClearCachePath string
	IgnoreNoExt bool
	IgnoreExt string
	ProvideApi bool
	ApiKey string
	CdnServerName string
	UrlMap string
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
	ProvideApi, err := conf.GetBool("default","provide_api")
	ApiKey, err := conf.GetString("default","api_key")
	CdnServerName, err := conf.GetString("default", "cdn_server_name")
	UrlMap, err := conf.GetString("default", "url_map")
	ApiStorePrefix, err := conf.GetString("default", "api_store_prefix")

	return &AnscdnConf{BaseServer,
		ServingPort, StoreDir, ApiStorePrefix, Strict,CacheOnly,FileMon,CacheExpires,
		ClearCachePath,IgnoreNoExt,IgnoreExt, ProvideApi,ApiKey,CdnServerName,UrlMap}, err
}
