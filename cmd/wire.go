package cmd

import (
	"fmt"

	"github.com/sanmoo/bruwrapper/internal/adapters/brucatalog"
	"github.com/sanmoo/bruwrapper/internal/adapters/brurunner"
	"github.com/sanmoo/bruwrapper/internal/adapters/interactive"
	"github.com/sanmoo/bruwrapper/internal/adapters/terminal"
	"github.com/sanmoo/bruwrapper/internal/adapters/yamlconfig"
	"github.com/sanmoo/bruwrapper/internal/core"
)

func wireUp() (core.Config, core.Catalog, core.Runner, core.Presenter, core.Selector, error) {
	cfgLoader := yamlconfig.New(yamlconfig.DefaultConfigPath())
	cfg, err := cfgLoader.Load()
	if err != nil {
		return core.Config{}, nil, nil, nil, nil, fmt.Errorf("loading config: %w\n\nCreate ~/.bruwrapper.yaml with your collection paths", err)
	}

	catalog := brucatalog.NewCatalog(cfg.CollectionPaths)

	bruPath, err := brurunner.FindBru()
	if err != nil {
		return core.Config{}, nil, nil, nil, nil, err
	}
	runner := brurunner.NewRunner(bruPath)

	presenter := terminal.NewPresenter(terminal.NewStdoutWriter())
	selector := interactive.NewSelector()

	return cfg, catalog, runner, presenter, selector, nil
}
