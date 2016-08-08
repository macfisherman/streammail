package main

import (
    "flag"
    "github.com/macfisherman/streammail/streamclient"
)

func main() {
    uri := flag.String("uri", "http://localhost:8080", "uri of the stream server")
    address := flag.String("address", "", "address to operate against")
    flag.Parse()
    _ = streamclient.NewStream(*uri, *address)
}