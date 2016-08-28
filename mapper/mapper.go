package mapper

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/charmap"
)

// MapValues show record from dictionary
func MapValues(tpl io.Reader, dict io.Reader) ([]string, error) {
	dec := charmap.Windows1251.NewDecoder()
	decr := dec.Reader(dict)
	csvr := csv.NewReader(decr)
	csvr.Comma = ';'

	var index int
	var dictNames []string
	dictionary := make(map[string]string)
	b, err := ioutil.ReadAll(tpl)

	if err != nil {
		fmt.Println(err)
	}

	pwd, err := os.Getwd()
	tmpBase := filepath.Join(pwd, "localtemp")
	var resFiles []string

	for {
		index++
		record, err := csvr.Read()
		if err != nil {
			break
		}
		if index == 1 {
			continue
		}
		if index == 2 {
			dictNames = record
			continue
		}
		for k, v := range dictNames {
			dictionary[v] = record[k]
		}

		tmpdir, err := ioutil.TempDir(tmpBase, "")
		defer os.RemoveAll(tmpdir)

		err = UnpackDocx(b, dictionary, tmpdir)
		if err != nil {
			return resFiles, err
		}
		fn := tmpdir + "application.docx"
		err = MakeDocx(tmpdir, fn)
		if err != nil {
			return resFiles, err
		}

		resFiles = append(resFiles, fn)
	}

	return resFiles, err
}

// MakeDocx zip files from source to target
func MakeDocx(source, target string) error {
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

		header.Name = filepath.Join("", strings.TrimPrefix(path, source+"\\"))
		header.Method = zip.Deflate

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
func UnpackDocx(b []byte, dict map[string]string, tmpdir string) error {
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
				exp, _ := regexp.Compile("%" + k + "%")
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
