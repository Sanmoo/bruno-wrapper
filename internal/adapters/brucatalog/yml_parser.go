package brucatalog

import (
	"fmt"
	"os"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/core"
	"gopkg.in/yaml.v3"
)

// ymlFileNew represents the new Bruno YML format
type ymlFileNew struct {
	Info ymlInfoNew `yaml:"info"`
	HTTP ymlHTTPNew `yaml:"http"`
	Body ymlBodyNew `yaml:"body"`
}

type ymlInfoNew struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Seq  int    `yaml:"seq"`
}

type ymlHTTPNew struct {
	Method  string         `yaml:"method"`
	URL     string         `yaml:"url"`
	Headers []ymlHeaderNew `yaml:"headers"`
}

type ymlHeaderNew struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type ymlBodyNew struct {
	Type string `yaml:"type"`
	Data string `yaml:"data"`
}

// ymlFileOld represents the old test format
type ymlFileOld struct {
	Meta    ymlMetaOld        `yaml:"meta"`
	HTTP    ymlHTTPOld        `yaml:"http"`
	Headers map[string]string `yaml:"headers"`
	Body    ymlBodyOld        `yaml:"body"`
}

type ymlMetaOld struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	Seq  int    `yaml:"seq"`
}

type ymlHTTPOld struct {
	Method string `yaml:"method"`
	URL    string `yaml:"url"`
}

type ymlBodyOld struct {
	Type    string `yaml:"type"`
	Content string `yaml:"content"`
}

func ParseYMLFile(path string) (core.Request, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return core.Request{}, fmt.Errorf("read yml file: %w", err)
	}

	// Try new format first (info.name, http.headers array, body.data)
	var newFile ymlFileNew
	if err := yaml.Unmarshal(data, &newFile); err == nil && newFile.Info.Name != "" {
		headers := make(map[string]string)
		for _, h := range newFile.HTTP.Headers {
			if h.Name != "" {
				headers[h.Name] = h.Value
			}
		}

		return core.Request{
			Name:    newFile.Info.Name,
			Method:  core.RequestMethod(strings.ToUpper(newFile.HTTP.Method)),
			URL:     newFile.HTTP.URL,
			Headers: headers,
			Body:    strings.TrimSpace(newFile.Body.Data),
			Path:    path,
		}, nil
	}

	// Fall back to old format (meta.name, headers map at root, body.content)
	var oldFile ymlFileOld
	if err := yaml.Unmarshal(data, &oldFile); err != nil {
		return core.Request{}, fmt.Errorf("parse yml file: %w", err)
	}

	if oldFile.Meta.Name == "" {
		return core.Request{}, fmt.Errorf("parse yml file: request name not found in either format")
	}

	headers := oldFile.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	return core.Request{
		Name:    oldFile.Meta.Name,
		Method:  core.RequestMethod(strings.ToUpper(oldFile.HTTP.Method)),
		URL:     oldFile.HTTP.URL,
		Headers: headers,
		Body:    strings.TrimSpace(oldFile.Body.Content),
		Path:    path,
	}, nil
}
