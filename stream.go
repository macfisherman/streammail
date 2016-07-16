package main

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
}

func Register(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	if (address[0] != 'S') || (address[0] != 'R') {
		fmt.Fprintf(w, "{ \"error\": \"address is not a STREAM address\" } \n")
		return
	}
	_, _, err := base58.CheckDecode(address)
	if err != nil {
		fmt.Fprintf(w, "{ \"error\": \"address format is invalid.\" }\n")
		return
	}

	err = os.Mkdir(ps.ByName("address"), 0755)
	if err != nil {
		fmt.Fprintf(w, "{ \"error\": \"unable to create address\"}")
	} else {
		fmt.Fprintf(w, "{ \"status\": \"address registered\"}")
	}
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/register/:address", Register)
	router.GET("/hello/:name", Hello)

	log.Fatal(http.ListenAndServe(":8080", router))
}
