package mapper

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"golang.org/x/text/encoding/charmap"
)

type Mapper interface {
	MapValues(io.Reader, io.Reader, string) (string, error)
}

type MapperCSVtoDOCX struct {
	tplName string
}

type MapperJSONtoDOCX struct {
	tplName string
}

type HelperDOCX struct {}

func (m MapperJSONtoDOCX) MapValues(tpl io.Reader, dict io.Reader, tplName string) (string, error) {
	helper := HelperDOCX{}
	var jsonMap []map[string]string
	dictBytes, err := ioutil.ReadAll(dict)
	err = json.Unmarshal(dictBytes, &jsonMap)
	tplBytes, err := ioutil.ReadAll(tpl)

	if err != nil {
		return "", err
	}

	return helper.GenerateArchiveDOCX(tplBytes, jsonMap, tplName)
}

// MapValues show record from dictionary
func (m MapperCSVtoDOCX) MapValues(tpl io.Reader, dict io.Reader, tplName string) (string, error) {
	helper := HelperDOCX{}
	dec := charmap.Windows1251.NewDecoder()
	decr := dec.Reader(dict)
	csvr := csv.NewReader(decr)
	csvr.Comma = ';'

	var index int
	var dictNames []string
	var dictionary []map[string]string

	for {
		index++
		row, err := csvr.Read()
		if err != nil {
			break
		}
		if index == 0 {
			continue
		}
		if index == 1 {
			dictNames = row
			continue
		}

		record := make(map[string]string)
		for k, v := range dictNames {
			record[v] = row[k]
		}

		dictionary = append(dictionary, record)
	}

	tplBytes, err := ioutil.ReadAll(tpl)
	if err != nil {
		return "", err
	}

	return helper.GenerateArchiveDOCX(tplBytes, dictionary, tplName)
}

func (helper HelperDOCX) GenerateArchiveDOCX(tpl []byte, dict []map[string]string, tplName string) (string, error) {
	tmpBase, err := ioutil.TempDir("", "operation")
	defer os.RemoveAll(tmpBase)
	var resFiles []string

	zipdir, err := ioutil.TempDir("", "zip")
	zipFilename := filepath.Join(zipdir, "result.zip")

	//fmt.Println(tmpBase, zipFilename, zipdir)

	newfile, err := os.Create(zipFilename)
	if err != nil {
		return "", err
	}
	defer newfile.Close()

  zwriter := zip.NewWriter(newfile)
	defer zwriter.Close()

	for key, record := range dict {
		tmpdir, err := ioutil.TempDir(tmpBase, "")
		err = helper.UnpackDocx(tpl, record, tmpdir)

		if err != nil {
			return "", err
		}

		var docName = "document"

		if len(tplName) > 0 {
			docName = tplName
		}

		var extension = filepath.Ext(docName)
		docName = docName[0:len(docName)-len(extension)]

		fn := filepath.Join(tmpBase, "#" + strconv.Itoa(key + 1) + "_" + docName + ".docx")

		//fmt.Println(tmpdir, fn)

		err = helper.GenerateSingleDocx(tmpdir, fn)
		if err != nil {
			return "", err
		}

		zippingFile, err := os.Open(fn)
		if err != nil {
			return "", err
		}
		defer os.RemoveAll(fn)
		defer zippingFile.Close()

		info, err := zippingFile.Stat()
		if err != nil {
			return "", err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return "", err
		}

		writer, err := zwriter.CreateHeader(header)
		if err != nil {
			return "", err
		}

		_, err = io.Copy(writer, zippingFile)

		resFiles = append(resFiles, fn)
	}

	return zipFilename, err
}

// MakeDocx zip files from source to target
func (h HelperDOCX) GenerateSingleDocx(source, target string) error {
	docxFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer docxFile.Close()

	archive := zip.NewWriter(docxFile)
	defer archive.Close()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(source, path)
		header.Method = zip.Deflate

		//fmt.Println("MY", header.Name, path, source)

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// UnpackDocx unzip and change files
func (h HelperDOCX) UnpackDocx(b []byte, dict map[string]string, tmpdir string) error {
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return fmt.Errorf("error unzipping data: %v", err)
	}

	var perm os.FileMode = 0777
	err = os.Mkdir(tmpdir+"/_rels", perm)
	err = os.Mkdir(tmpdir+"/word", perm)
	err = os.Mkdir(tmpdir+"/word/_rels", perm)
	err = os.Mkdir(tmpdir+"/word/theme", perm)
	err = os.Mkdir(tmpdir+"/docProps", perm)

	for _, f := range zr.File {
		fr, err := f.Open()
		defer fr.Close()
		if err != nil {
			return fmt.Errorf("error opening '%v' from archive: %v", f.Name, err)
		}

		newName := filepath.Join(tmpdir, f.Name)

		if f.Name == "word/document.xml" {
			b, err := ioutil.ReadAll(fr)
			if err != nil {
				return err
			}

			s := string(b)
			for k, v := range dict {
				exp, _ := regexp.Compile("x" + k + "x")
				indexes := exp.FindStringIndex(s)
				if indexes == nil {
					err = fmt.Errorf(k)
				}
				s = exp.ReplaceAllLiteralString(s, v)
			}

			ioutil.WriteFile(newName, []byte(s), 0777)
		} else {
			out, err := os.Create(newName)
			defer out.Close()
			if err != nil {
				return fmt.Errorf("error creating file: %v", err)
			}

			_, err = io.Copy(out, fr)
			if err != nil {
				return fmt.Errorf("error copy file: %v", err)
			}

		}
	}

	return err
}
