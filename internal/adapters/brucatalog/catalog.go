package brucatalog

import (
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
		name, err := detectCollection(expanded)
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

	_, err = detectCollection(colPath)
	if err != nil {
		return nil, err
	}

	var requests []core.Request

	err = filepath.WalkDir(colPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".yml") {
			return nil
		}

		req, err := ParseYMLFile(path)
		if err != nil {
			return nil
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

func detectCollection(dirPath string) (string, error) {
	ymlPath := filepath.Join(dirPath, "opencollection.yml")
	if info, err := os.Stat(ymlPath); err == nil && !info.IsDir() {
		data, err := os.ReadFile(ymlPath)
		if err != nil {
			return "", err
		}

		// Try new format first (opencollection: x.x.x, info.name: ...)
		var ocNew openCollectionYMLNew
		if err := yaml.Unmarshal(data, &ocNew); err == nil && ocNew.Info.Name != "" {
			return ocNew.Info.Name, nil
		}

		// Fall back to old format (version: "x", name: ...)
		var ocOld openCollectionYMLOld
		if err := yaml.Unmarshal(data, &ocOld); err != nil {
			return "", err
		}
		if ocOld.Name != "" {
			return ocOld.Name, nil
		}
		return "", fmt.Errorf("opencollection.yml: collection name not found")
	}

	return "", fmt.Errorf("no opencollection.yml found in %q", dirPath)
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
		name, err := detectCollection(expanded)
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
