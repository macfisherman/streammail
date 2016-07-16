package main

import (
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
)

func report_error(w http.ResponseWriter, e string) {
	fmt.Fprintf(w, "{ \"error\": \"%s\" } \n", e)
}

func report_status(w http.ResponseWriter, s string) {
	fmt.Fprintf(w, "{ \"status\": \"%s\" } \n", s)
}

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
		report_error(w, "address format is invalid")
		return
	}

	err = os.Mkdir(ps.ByName("address"), 0755)
	if err != nil {
		report_error(w, "unable to create address:"+err.Error())
	} else {
		report_status(w, "address registered")
	}
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.POST("/register/:address", Register)
	router.GET("/hello/:name", Hello)

	log.Fatal(http.ListenAndServe(":8080", router))
}
