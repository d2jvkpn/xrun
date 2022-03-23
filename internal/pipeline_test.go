package internal

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"
)

func TestTemplate(t *testing.T) {
	templ, err := template.New("test").Parse(`{{.name}} {{.x}}`)
	if err != nil {
		t.Fatal(err)
	}

	data := map[string]string{
		"name": "rover",
		"x":    "2022",
	}

	buf := new(bytes.Buffer)
	if err = templ.Execute(buf, data); err != nil {
		t.Fatal(err)
	}

	fmt.Println(buf.String())
}

func TestLoadPipeline(t *testing.T) {
	p, err := LoadPipeline("../examples/pipeline.yaml")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", p)

	fmt.Printf("%+v\n", p.Tasks[0].commands)

	errs := p.run(0, -1)
	fmt.Println(errs)
}
