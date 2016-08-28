package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/usehotkey/mapper/mapper"
)

func runMapper(w http.ResponseWriter, r *http.Request) {
	var (
		status    int
		err       error
		tpl, dict io.Reader
	)
	defer func() {
		if nil != err {
			http.Error(w, err.Error(), status)
		}
	}()
	const _24K = (1 << 20) * 24
	if err = r.ParseMultipartForm(_24K); nil != err {
		status = http.StatusInternalServerError
		return
	}
	for name, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			var infile multipart.File
			if infile, err = hdr.Open(); nil != err {
				status = http.StatusInternalServerError
				return
			}

			if name == "tpl" {
				tpl = infile
			}

			if name == "dict" {
				dict = infile
			}
		}
	}

	err = mapper.MapValues(tpl, dict)
	if err != nil {
		fmt.Println(err)
	}
}

func mainPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	b, err := ioutil.ReadFile("index.html")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, string(b))
}

func main() {
	http.HandleFunc("/run", runMapper)
	http.HandleFunc("/", mainPage)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
