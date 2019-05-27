package webgen

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path"
	"strings"

	"github.com/winded/go-webgen/rel_io"
)

type StaticFileGenerator struct {
	inputReader  rel_io.IFileReader
	outputWriter rel_io.IFileWriter
	ignoredFiles map[string]struct{}
	ioMap        map[string]string
}

func NewStaticFileGenerator(input rel_io.IFileReader, output rel_io.IFileWriter) *StaticFileGenerator {
	return &StaticFileGenerator{input, output, make(map[string]struct{}), nil}
}

func (this *StaticFileGenerator) Ignore(files ...string) {
	for _, f := range files {
		this.ignoredFiles[f] = struct{}{}
	}
}

// GetFile returns the output path of the given input path file. If GenerateMap has not been called, a panic will occur
func (this *StaticFileGenerator) GetFile(inputPath string) string {
	if this.ioMap == nil {
		panic(errors.New("No IO map found. Use GenerateMap first"))
	}

	return this.ioMap[inputPath]
}

func (this *StaticFileGenerator) GenerateMap() error {
	files, err := this.inputReader.List()
	if err != nil {
		return err
	}

	ioMap := make(map[string]string)

	hasher := sha256.New()
	for _, f := range files {
		if _, found := this.ignoredFiles[f]; found {
			continue
		}

		data, err := this.inputReader.Read(f)
		if err != nil {
			return err
		}

		ext := path.Ext(f)
		name := path.Base(f)
		name = name[:strings.LastIndex(name, ext)]

		_, err = hasher.Write(data)
		if err != nil {
			return err
		}
		hash := hex.EncodeToString(hasher.Sum(nil))
		hasher.Reset()

		ioMap[f] = "/" + name + "_" + hash + ext
	}

	this.ioMap = ioMap
	return nil
}

func (this *StaticFileGenerator) Generate(useGzip bool) error {
	if this.ioMap == nil {
		err := this.GenerateMap()
		if err != nil {
			return err
		}
	}

	gz := gzip.NewWriter(nil)
	for iFile, oFile := range this.ioMap {
		data, err := this.inputReader.Read(iFile)
		if err != nil {
			return err
		}

		err = this.outputWriter.Write(oFile, data)
		if err != nil {
			return err
		}

		if useGzip {
			buf := bytes.NewBuffer(make([]byte, 0))
			gz.Reset(buf)
			_, err = gz.Write(data)
			if err != nil {
				return err
			}
			err = gz.Close()
			if err != nil {
				return err
			}
			err = this.outputWriter.Write(oFile+".gz", buf.Bytes())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
