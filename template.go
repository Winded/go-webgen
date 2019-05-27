package webgen

import (
	"bytes"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/winded/go-webgen/rel_io"
)

type TemplateFileGenerator struct {
	t            *template.Template
	inputReader  rel_io.IFileReader
	outputWriter rel_io.IFileWriter
}

func NewTemplateFileGenerator(input rel_io.IFileReader, output rel_io.IFileWriter) *TemplateFileGenerator {
	return &TemplateFileGenerator{
		inputReader:  input,
		outputWriter: output,
		t:            template.New("root"),
	}
}

func (this *TemplateFileGenerator) Funcs(fmap template.FuncMap) {
	this.t.Funcs(fmap)
}

func (this *TemplateFileGenerator) AddTemplates(pathMatch *regexp.Regexp) error {
	files, err := this.inputReader.List()
	if err != nil {
		return err
	}

	for _, f := range files {
		if !pathMatch.MatchString(f) {
			continue
		}

		name := path.Base(f)
		name = name[:strings.LastIndex(name, path.Ext(f))]
		err = this.AddTemplate(name, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *TemplateFileGenerator) AddTemplate(name, path string) error {
	data, err := this.inputReader.Read(path)
	if err != nil {
		return err
	}

	_, err = this.t.New(name).Parse(string(data))
	if err != nil {
		return err
	}

	return nil
}

func (this *TemplateFileGenerator) Generate(outputPath, templateName string, data interface{}) error {
	d, err := this.GenerateBytes(templateName, data)
	if err != nil {
		return err
	}

	err = this.outputWriter.Write(outputPath, d)
	if err != nil {
		return err
	}

	return nil
}

func (this *TemplateFileGenerator) GenerateBytes(templateName string, data interface{}) ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1024))
	err := this.t.ExecuteTemplate(buffer, templateName, data)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
