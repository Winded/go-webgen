package main

import (
	"fmt"
	"os"
	"path"

	"github.com/winded/go-webgen"
)

type TemplateData struct {
	List []string
}

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	gen := webgen.NewGenerator(webgen.GeneratorConfig{
		OutputDir:   path.Join(rootDir, "output"),
		TemplateDir: path.Join(rootDir, "templates"),
		StaticDir:   path.Join(rootDir, "static"),

		CompressStaticFiles: true,
		StaticOutputPrefix:  "/static",
	})

	gen.Add("/index.html", "index", &TemplateData{
		List: []string{
			"Space",
			"Mind",
			"Reality",
			"Power",
			"Time",
			"Soul",
		},
	})

	if err := gen.Generate(); err != nil {
		panic(err)
	}

	fmt.Println("Generation successful")
}
