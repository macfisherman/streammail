package main

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

func WriteJSON(w http.ResponseWriter, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(b)
	return nil
}

func report_error(w http.ResponseWriter, code int, err string) {
	w.WriteHeader(code)

	WriteJSON(w, map[string]string{
		"error": err,
	})
}

func report_status(w http.ResponseWriter, code int, v interface{}) error {
	w.WriteHeader(code)

	return WriteJSON(w, v)
}

func IndexPage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
}

func PostMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	filename := time.Now().UTC().Format(time.RFC3339Nano)
	path := address + "/" + filename
	msg, err := os.Create(path)
	if err != nil {
		report_error(w, 409, "in creating message file "+path+": "+err.Error())
		return
	}
	defer msg.Close()

	if _, err := io.Copy(msg, r.Body); err != nil {
		report_error(w, 409, "in serializing messagee "+path+": "+err.Error())
		return
	}

	w.Header().Set("Location", "/message/"+path)
	report_status(w, 201, map[string]string{"ok": "saved " + path})
}

func GetMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	filepath := ps.ByName("address") + "/" + ps.ByName("id")
	msg, err := os.Open(filepath)
	if err != nil {
		report_error(w, 409, filepath+": "+err.Error())
		return
	}
	defer msg.Close()

	// might want to rethink how msg is just a blob and not a JSON object
	if _, err := io.Copy(w, msg); err != nil {
		report_error(w, 409, err.Error())
		return
	}
}

func Index(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	address := ps.ByName("address")
	vars := r.URL.Query()

	dirHandle, err := os.Open(address)
	if err != nil {
		report_error(w, 404, err.Error())
		return
	}
	defer dirHandle.Close()

	files, err := dirHandle.Readdir(0)
	if err != nil {
		report_error(w, 409, err.Error())
		return
	}

	var names []string
	for _, file := range files {
		if file.Mode().IsRegular() {
			names = append(names, file.Name())
		}
	}

	sort.Strings(names)

	// setup a count - default to 100
	count := 100
	skipTo := vars.Get("from")
	if n := vars.Get("count"); n != "" {
		count, err = strconv.Atoi(n)
		if err != nil {
			report_error(w, 400, "invalid number "+n+" :"+err.Error())
			return
		}
	}

	have := 0
	if skipTo != "" {
		var wantedNames []string
		getRemaining := false
		for _, name := range names {
			if name == skipTo {
				getRemaining = true
			}
			if getRemaining {
				wantedNames = append(wantedNames, name)
				have++
				if have == count {
					break
				}
			}
		}

		names = wantedNames
	}

	if count > len(names) {
		count = len(names)
	}
	encoder := json.NewEncoder(w)
	err = encoder.Encode(names[:count]) // only return count
	if err != nil {
		report_error(w, 409, err.Error())
		return
	}
}

type Data struct {
	Address string
}

func Register(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var result Data
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		r.Body.Close()
		report_error(w, 400, "unable to parse JSON: "+err.Error())
	}
	address := result.Address
	log.Printf("address is: %c:%v", address[0], address)
	if !((address[0] == 'S') || (address[0] == 'R')) {
		report_error(w, 400, "address not a STREAM address")
		return
	}
	if _, _, err := base58.CheckDecode(address); err != nil {
		report_error(w, 400, "address format is invalid")
		return
	}

	if err := os.Mkdir(address, 0755); err != nil {
		report_error(w, 409, "unable to create address:"+err.Error())
	} else {
		w.Header().Set("Location", "/address/"+address)
		report_status(w, 201, map[string]string{"ok": "address registered"})
	}
}

func main() {
	router := httprouter.New()
	router.GET("/", IndexPage)
	router.POST("/address", Register)
	router.POST("/message/:address", PostMessage)
	router.GET("/index/:address", Index)
	router.GET("/message/:address/:id", GetMessage)

	n := negroni.Classic()
	n.UseHandler(router)

	log.Fatal(http.ListenAndServe(":8080", n))
}
