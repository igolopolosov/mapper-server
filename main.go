package main

import (
	"os"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"net/http"
	"strings"

	"github.com/usehotkey/mapper/mapper"
)

func runCSVtoDOCX(w http.ResponseWriter, r *http.Request) {
	var (
		status    int
		err       error
		tpl, dict io.Reader
		tplType, dictType string
		m mapper.Mapper
	)

	defer func() {
		if nil != err {
			http.Error(w, err.Error(), status)
		}
	}()
	const _24K = (1 << 20) * 24
	if err = r.ParseMultipartForm(_24K); err != nil  {
		http.Error(w, err.Error(), 500)
		return
	}
	for name, fheaders := range r.MultipartForm.File {
		for _, hdr := range fheaders {
			var infile multipart.File
			if infile, err = hdr.Open(); err != nil {
				http.Error(w, err.Error(), 500)
		    return
			}

			if name == "tpl" && strings.Contains(hdr.Filename, ".docx") {
				tpl = infile
				tplType = "DOCX"
			}

			if name == "dict" && strings.Contains(hdr.Filename, ".csv") {
				dict = infile
				dictType = "CSV"
			}

			if name == "dict" && strings.Contains(hdr.Filename, ".json") {
				dict = infile
				dictType = "JSON"
			}
		}
	}

	if (tpl == nil || dict == nil) {
		http.Error(w, "Incorrect file types", 500)
    return
	}

	w.Header().Set("Content-Type", "text/html")

	if tplType == "DOCX" && dictType == "CSV" {
		m = mapper.MapperCSVtoDOCX{}
	}

	if tplType == "DOCX" && dictType == "JSON" {
		m = mapper.MapperJSONtoDOCX{}
	}

	zipName, err := m.MapValues(tpl, dict)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=" + filepath.Base(zipName))
	http.ServeFile(w, r, zipName)
	defer os.RemoveAll(zipName)
}

func main() {
	log.Print("Start on:" + os.Getenv("PORT"))
	http.HandleFunc("/map", runCSVtoDOCX)
	err := http.ListenAndServe(":" + os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal("ListenAndServe:" + os.Getenv("PORT"), err)
	}
}
