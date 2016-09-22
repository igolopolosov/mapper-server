package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/usehotkey/mapper/mapper"
)

func runCSVtoDOCX(w http.ResponseWriter, r *http.Request) {
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
	if err = r.ParseMultipartForm(_24K); err != nil  {
		status = http.StatusInternalServerError
		return
	}
	for name, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			var infile multipart.File
			if infile, err = hdr.Open(); err != nil {
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

	m := mapper.MapperCSVtoDOCX{}

	files, err := m.MapValues(tpl, dict)
	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "text/html")
	for k, v := range files {
		fmt.Fprintf(w, "#%v %v \n", k+1, v)
	}

}

func main() {
	log.Print("Start")
	http.HandleFunc("/csvtodocx", runCSVtoDOCX)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe:9090 ", err)
	}
}
