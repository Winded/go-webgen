// rel_io contains abstraction interfaces for reading and writing data to a relative path.
package rel_io

import (
	"io/ioutil"
	"os"
	"path"
)

type IFileReader interface {
	List() ([]string, error)
	Read(path string) ([]byte, error)
	Copy(writer IFileWriter, from, to string) error
}

type IFileWriter interface {
	Write(path string, data []byte) error
}

type IFileReadWriter interface {
	IFileReader
	IFileWriter
}

type StandardFileReadWriter struct {
	rootDir string
}

func NewStandardFileReadWriter(rootDir string) *StandardFileReadWriter {
	return &StandardFileReadWriter{
		rootDir: rootDir,
	}
}

func (this *StandardFileReadWriter) List() ([]string, error) {
	var walk func(string, func(string))
	walk = func(root string, callback func(string)) {
		files, _ := ioutil.ReadDir(root)
		for _, f := range files {
			if f.IsDir() {
				walk(root+"/"+f.Name(), callback)
			} else {
				callback(root + "/" + f.Name())
			}
		}
	}

	list := make([]string, 0, 5)
	walk(this.rootDir, func(path string) {
		list = append(list, path[len(this.rootDir):])
	})

	return list, nil
}

func (this *StandardFileReadWriter) Read(path string) ([]byte, error) {
	realPath := this.rootDir + path

	stat, err := os.Stat(realPath)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(realPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data := make([]byte, stat.Size())
	_, err = f.Read(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (this *StandardFileReadWriter) Copy(writer IFileWriter, from, to string) error {
	data, err := this.Read(from)
	if err != nil {
		return err
	}

	return writer.Write(to, data)
}

func (this *StandardFileReadWriter) Write(relativePath string, data []byte) error {
	realPath := this.rootDir + relativePath

	err := os.MkdirAll(path.Dir(realPath), 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(realPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}
