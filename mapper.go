package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MakeAppl makes populated docx file
func MakeAppl(tpl io.Reader) (string, error) {
	dictionary := map[string]string{
		"DATE": "13.06.1994",
		"NAME": "Igor",
	}

	pwd, err := os.Getwd()
	tmpBase := filepath.Join(pwd, "localtemp")
	tmpdir, err := ioutil.TempDir(tmpBase, "")
	defer os.RemoveAll(tmpdir)

	err = UnpackDocx(tpl, dictionary, tmpdir)
	fn := tmpdir + "application.docx"
	err = MakeDocx(tmpdir, fn)

	if err != nil {
		return fmt.Sprintf("MakeAppl: %v", err), err
	}
	return fn, err
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
func UnpackDocx(r io.Reader, dict map[string]string, tmpdir string) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("error reading data: %v", err)
	}
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
			err = ChangeValues(fr, dict, newName)
			if err != nil {
				return fmt.Errorf("error changing values: %v", err)
			}
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

// ChangeValues in file by map
func ChangeValues(r io.Reader, m map[string]string, outfn string) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	s := string(b)
	for k, v := range m {
		regexp, _ := regexp.Compile("%" + k + "%")
		s = regexp.ReplaceAllLiteralString(s, v)
	}

	ioutil.WriteFile(outfn, []byte(s), 0777)

	return err
}
