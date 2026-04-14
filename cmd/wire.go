package cmd

import (
	"fmt"
	"os"

	"github.com/sanmoo/bruwrapper/internal/adapters/brucatalog"
	"github.com/sanmoo/bruwrapper/internal/adapters/brurunner"
	"github.com/sanmoo/bruwrapper/internal/adapters/interactive"
	"github.com/sanmoo/bruwrapper/internal/adapters/terminal"
	"github.com/sanmoo/bruwrapper/internal/adapters/yamlconfig"
	"github.com/sanmoo/bruwrapper/internal/core"
)

func resolveConfigPath() string {
	path := cfgPath
	if path == "" {
		path = os.Getenv("BRUWRAPPER_CONFIG")
	}
	if path == "" {
		path = yamlconfig.DefaultConfigPath()
	}
	return path
}

func wireCatalogAndPresenter() (core.Catalog, core.Presenter, error) {
	cfgLoader := yamlconfig.New(resolveConfigPath())
	cfg, err := cfgLoader.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w\n\nCreate ~/.bruwrapper.yaml with your collection paths", err)
	}
	catalog := brucatalog.NewCatalog(cfg.CollectionPaths)
	presenter := terminal.NewPresenter(terminal.NewStdoutWriter())
	return catalog, presenter, nil
}

func wireUp() (core.Catalog, core.Runner, core.Presenter, core.Selector, error) {
	catalog, presenter, err := wireCatalogAndPresenter()
	if err != nil {
		return nil, nil, nil, nil, err
	}

	bruPath, err := brurunner.FindBru()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	runner := brurunner.NewRunner(bruPath)
	selector := interactive.NewSelector()

	return catalog, runner, presenter, selector, nil
}
