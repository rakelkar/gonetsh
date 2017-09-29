# gonetsh

A simple set of GO functions to wrap windows netsh commands. Inspired by the netsh wrapper in kubernetes. Now also provides netroute that wraps route CRUD powershell commandlets.

## Build
`go build ./...`

## Test
`go test ./...`

## Integration Test
Integration tests actually runs netsh on your machine...
`go test -tags=integration -v`