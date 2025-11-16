@echo off
setlocal
pushd "%~dp0"
set GOTELEMETRY=off
set GOAMD64=v3
set GOEXPERIMENT=greenteagc
go build -trimpath -gcflags="all=-B -C -dwarf=false" -ldflags="-s -w -buildid=" -buildvcs=false
popd
