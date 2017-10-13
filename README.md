# gonetsh
![build status](https://ci.appveyor.com/api/projects/status/32r7s2skrgm9ubva?svg=true)

A simple set of GO functions to wrap windows netsh commands. Inspired by the netsh wrapper in kubernetes. Now also provides netroute that wraps route CRUD powershell commandlets.

## Build
`./build.ps1`

## Test
`./test.ps1`

## Integration Test
Integration tests actually runs netsh on your machine...
```$bash
go test -tags=integration -v ./netroute
go test -tags=integration -v ./netsh
```
