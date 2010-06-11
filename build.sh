#!/bin/sh

$GOBIN/6g filemon.go
$GOBIN/6g anlog.go
$GOBIN/6g config.go
$GOBIN/6g anscdn.go
$GOBIN/6l -o anscdn anscdn.6

