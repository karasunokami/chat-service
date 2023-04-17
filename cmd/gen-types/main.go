package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 5 {
		log.Fatalf("invalid args count: %d", len(os.Args)-1)
	}

	pkg, types, out, tmplPath := os.Args[1], strings.Split(os.Args[2], ","), os.Args[3], os.Args[4]
	if err := run(pkg, types, out, tmplPath); err != nil {
		log.Fatal(err)
	}

	p, _ := os.Getwd()
	log.Printf("%v generated\n", filepath.Join(p, out))
}

func run(pkg string, types []string, outFile, tmplPath string) error {
	content, err := generateFileContent(tmplPath, types, pkg)
	if err != nil {
		return fmt.Errorf("generate file content, err=%w", err)
	}

	file, err := os.OpenFile(outFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return fmt.Errorf("open file, err=%v", err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("write string to file, err=%v", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("close file, err=%v", err)
	}

	return nil
}

func generateFileContent(tmplPath string, types []string, pkg string) (string, error) {
	tmplBytes, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("read tmpl file, err=%v", err)
	}

	buf := bytes.NewBufferString("")

	tpl := template.New("types").Funcs(template.FuncMap{"StringsJoin": strings.Join})
	tpl, err = tpl.Parse(string(tmplBytes))
	if err != nil {
		return "", fmt.Errorf("parse types template, err=%v", err)
	}

	err = tpl.Execute(buf, struct {
		Types   []string
		Package string
	}{
		Types:   types,
		Package: pkg,
	})
	if err != nil {
		return "", fmt.Errorf("execute template, err=%v", err)
	}

	return buf.String(), nil
}
