@echo off
setlocal
pushd "%~dp0"
set "GOAMD64=v3"
go build -gcflags="all=-B -C -dwarf=false" -ldflags="-s -w"