package brucatalog

import (
	"fmt"
	"os"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/core"
	"gopkg.in/yaml.v3"
)

type ymlFile struct {
	Meta    ymlMeta           `yaml:"meta"`
	HTTP    ymlHTTP           `yaml:"http"`
	Headers map[string]string `yaml:"headers"`
	Body    ymlBody           `yaml:"body"`
}

type ymlMeta struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Seq  int    `yaml:"seq"`
}

type ymlHTTP struct {
	Method string `yaml:"method"`
	URL    string `yaml:"url"`
}

type ymlBody struct {
	Type    string `yaml:"type"`
	Content string `yaml:"content"`
}

func ParseYMLFile(path string) (core.Request, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return core.Request{}, fmt.Errorf("read yml file: %w", err)
	}

	var f ymlFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return core.Request{}, fmt.Errorf("parse yml file: %w", err)
	}

	headers := f.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	req := core.Request{
		Name:    f.Meta.Name,
		Method:  core.RequestMethod(strings.ToUpper(f.HTTP.Method)),
		URL:     f.HTTP.URL,
		Headers: headers,
		Body:    strings.TrimSpace(f.Body.Content),
		Path:    path,
	}

	return req, nil
}
