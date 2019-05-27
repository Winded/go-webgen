package webgen

import (
	"path"
	"regexp"
	"text/template"

	"github.com/winded/go-webgen/rel_io"
)

var (
	defaultTemplatesRegexp *regexp.Regexp
)

func init() {
	defaultTemplatesRegexp = regexp.MustCompile("^/.+\\.html$")
}

type templateDataBind struct {
	Template string
	Data     interface{}
}

type GeneratorConfig struct {
	OutputDir   string
	TemplateDir string
	StaticDir   string

	URLPrefix           string
	StaticOutputPrefix  string
	CompressStaticFiles bool
	TemplatesRegexp     *regexp.Regexp
}

type Generator struct {
	staticGenerator   *StaticFileGenerator
	templateGenerator *TemplateFileGenerator

	config GeneratorConfig

	genMappings map[string]templateDataBind
}

func NewGenerator(config GeneratorConfig) *Generator {
	g := &Generator{}

	if config.OutputDir == "" {
		panic("OutputDir not set")
	}

	if config.StaticDir != "" {
		outputPrefix := config.StaticOutputPrefix
		if outputPrefix == "" {
			outputPrefix = "/static"
		}

		g.staticGenerator = NewStaticFileGenerator(rel_io.NewStandardFileReadWriter(config.StaticDir), rel_io.NewStandardFileReadWriter(path.Join(config.OutputDir, outputPrefix)))
	}

	if config.TemplateDir != "" {
		g.templateGenerator = NewTemplateFileGenerator(rel_io.NewStandardFileReadWriter(config.TemplateDir), rel_io.NewStandardFileReadWriter(config.OutputDir))

		g.templateGenerator.Funcs(map[string]interface{}{
			"static": g.getStatic,
		})
	}

	if config.TemplatesRegexp == nil {
		config.TemplatesRegexp = defaultTemplatesRegexp
	}

	g.config = config
	g.genMappings = make(map[string]templateDataBind)

	return g
}

func (this *Generator) Add(path, template string, data interface{}) {
	this.genMappings[path] = templateDataBind{
		Template: template,
		Data:     data,
	}
}

func (this *Generator) Funcs(fmap template.FuncMap) {
	this.templateGenerator.Funcs(fmap)
}

func (this *Generator) Generate() error {
	if this.staticGenerator != nil {
		if err := this.staticGenerator.GenerateMap(); err != nil {
			return err
		}
		if err := this.staticGenerator.Generate(this.config.CompressStaticFiles); err != nil {
			return err
		}
	}

	if this.templateGenerator != nil {
		if err := this.templateGenerator.AddTemplates(this.config.TemplatesRegexp); err != nil {
			return err
		}

		for path, td := range this.genMappings {
			if err := this.templateGenerator.Generate(path, td.Template, td.Data); err != nil {
				return err
			}
		}
	}

	return nil
}

func (this *Generator) getStatic(path string) string {
	if this.staticGenerator != nil {
		return this.config.URLPrefix + this.config.StaticOutputPrefix + this.staticGenerator.GetFile(path)
	} else {
		return ""
	}
}
