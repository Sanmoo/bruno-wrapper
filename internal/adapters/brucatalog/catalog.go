package brucatalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/core"
	"gopkg.in/yaml.v3"
)

type catalog struct {
	paths []string
}

type brunoJSON struct {
	Version string `json:"version"`
	Name    string `json:"name"`
	Type    string `json:"type"`
}

// openCollectionYMLNew represents the new Bruno opencollection.yml format
type openCollectionYMLNew struct {
	Opencollection string                 `yaml:"opencollection"`
	Info           openCollectionInfoNew  `yaml:"info"`
	Config         map[string]interface{} `yaml:"config,omitempty"`
	Bundled        bool                   `yaml:"bundled"`
	Extensions     map[string]interface{} `yaml:"extensions,omitempty"`
}

type openCollectionInfoNew struct {
	Name string `yaml:"name"`
}

// openCollectionYMLOld represents the old test format
type openCollectionYMLOld struct {
	Version string `yaml:"version"`
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
}

func NewCatalog(paths []string) core.Catalog {
	return &catalog{paths: paths}
}

func (c *catalog) FindCollections() ([]core.Collection, error) {
	var collections []core.Collection
	for _, p := range c.paths {
		expanded, err := expandPath(p)
		if err != nil {
			continue
		}
		name, _, err := detectCollection(expanded)
		if err != nil {
			continue
		}
		absPath, _ := filepath.Abs(expanded)
		collections = append(collections, core.Collection{
			Name: name,
			Path: absPath,
		})
	}
	return collections, nil
}

func (c *catalog) FindRequests(collectionName string) ([]core.Request, error) {
	colPath, err := c.findCollectionPath(collectionName)
	if err != nil {
		return nil, err
	}

	_, format, err := detectCollection(colPath)
	if err != nil {
		return nil, err
	}

	var requests []core.Request
	ext := ".bru"
	if format == core.FormatYML {
		ext = ".yml"
	}

	err = filepath.WalkDir(colPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ext) {
			return nil
		}

		var req core.Request
		if format == core.FormatBru {
			req, err = ParseBruFile(path)
			if err != nil {
				return nil
			}
		} else {
			req, err = ParseYMLFile(path)
			if err != nil {
				return nil
			}
		}
		req.Collection = collectionName
		requests = append(requests, req)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return requests, nil
}

func (c *catalog) ResolveRequest(collectionName, requestName string) (core.Request, error) {
	reqs, err := c.FindRequests(collectionName)
	if err != nil {
		return core.Request{}, err
	}
	for _, r := range reqs {
		if r.Name == requestName {
			return r, nil
		}
	}
	return core.Request{}, fmt.Errorf("request %q not found in collection %q", requestName, collectionName)
}

func detectCollection(dirPath string) (string, core.CollectionFormat, error) {
	brunoPath := filepath.Join(dirPath, "bruno.json")
	if info, err := os.Stat(brunoPath); err == nil && !info.IsDir() {
		data, err := os.ReadFile(brunoPath)
		if err != nil {
			return "", "", err
		}
		var bj brunoJSON
		if err := json.Unmarshal(data, &bj); err != nil {
			return "", "", err
		}
		return bj.Name, core.FormatBru, nil
	}

	ymlPath := filepath.Join(dirPath, "opencollection.yml")
	if info, err := os.Stat(ymlPath); err == nil && !info.IsDir() {
		data, err := os.ReadFile(ymlPath)
		if err != nil {
			return "", "", err
		}

		// Try new format first (opencollection: x.x.x, info.name: ...)
		var ocNew openCollectionYMLNew
		if err := yaml.Unmarshal(data, &ocNew); err == nil && ocNew.Info.Name != "" {
			return ocNew.Info.Name, core.FormatYML, nil
		}

		// Fall back to old format (version: "x", name: ...)
		var ocOld openCollectionYMLOld
		if err := yaml.Unmarshal(data, &ocOld); err != nil {
			return "", "", err
		}
		if ocOld.Name != "" {
			return ocOld.Name, core.FormatYML, nil
		}
		return "", "", fmt.Errorf("opencollection.yml: collection name not found")
	}

	return "", "", fmt.Errorf("no bruno.json or opencollection.yml found in %q", dirPath)
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}

func (c *catalog) findCollectionPath(collectionName string) (string, error) {
	for _, p := range c.paths {
		expanded, err := expandPath(p)
		if err != nil {
			continue
		}
		name, _, err := detectCollection(expanded)
		if err != nil {
			continue
		}
		if name == collectionName {
			absPath, _ := filepath.Abs(expanded)
			return absPath, nil
		}
	}
	return "", fmt.Errorf("collection %q not found", collectionName)
}
