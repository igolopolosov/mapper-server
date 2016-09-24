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

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "origin, x-requested-with, content-type")
	w.Header().Set("Access-Control-Allow-Methods", "POST")

	if r.Method == "OPTIONS" {
		return
	}

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
		http.Error(w, "Некорретный тип прикреплённых файлов", 500)
    return
	}

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

// func main() {
// 	log.Print("Start on:9090")
// 	http.HandleFunc("/map", runCSVtoDOCX)
// 	err := http.ListenAndServe(":9090", nil)
// 	if err != nil {
// 		log.Fatal("ListenAndServe:9090", err)
// 	}
// }
