package main

import (
    "flag"
    "fmt"
    "github.com/macfisherman/streammail/streamclient"
)

func main() {
    uri := flag.String("uri", "http://localhost:8080", "uri of the stream server")
    address := flag.String("address", "", "address to operate against")
    flag.Parse()

    // connect to server and get index
    stream := streamclient.NewStream(*uri, *address)
    _, err := stream.GetIndex()
    if err != nil {
        fmt.Println(err)
    }
}