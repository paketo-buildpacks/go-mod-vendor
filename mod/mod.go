package mod

import (
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const (
	Dependency = "go-mod"
	Launch     = "go-binary"
)

type Runner interface {
	Run(bin, dir string, quiet bool, args ...string) error
	RunWithOutput(bin, dir string, quiet bool, args ...string) (string, error)
	SetEnv(variableName string, path string) error
	Rename(existingPath string, newPath string) error
}

type Logger interface {
	Info(format string, args ...interface{})
}

type PkgManager interface {
	Install(location, cacheDir string) error
}

type MetadataInterface interface {
	Identity() (name string, version string)
}

type Metadata struct {
	Name string
	Hash string
}

func (m Metadata) Identity() (name string, version string) {
	return m.Name, m.Hash
}

type Contributor struct {
	goModMetadata MetadataInterface
	goModLayer    layers.Layer
	launchLayer   layers.Layer
	runner        Runner
	appRoot       string
	logger        Logger
	launch        layers.Layers
}

func NewContributor(context build.Build, runner Runner) (Contributor, bool, error) {
	_, wantDependency := context.BuildPlan[Dependency]
	if !wantDependency {
		return Contributor{}, false, nil
	}

	contributor := Contributor{
		goModLayer:    context.Layers.Layer(Dependency),
		launchLayer:   context.Layers.Layer(Launch),
		goModMetadata: nil,
		runner:        runner,
		appRoot:       context.Application.Root,
		logger:        context.Logger,
		launch:        context.Layers,
	}

	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	if err := c.goModLayer.Contribute(c.goModMetadata, c.contributeGoModules, []layers.Flag{layers.Cache}...); err != nil {
		return err
	}

	if err := c.Install(); err != nil {
		return err
	}

	if err := c.launchLayer.Contribute(c.goModMetadata, c.contributeGoModules, []layers.Flag{layers.Launch}...); err != nil {
		return err
	}

	return c.setStartCommand()
}

func (c Contributor) contributeGoModules(layer layers.Layer) error {
	return nil
}

func (c Contributor) Install() error {
	c.logger.Info("Setting environment variables")
	if err := c.runner.SetEnv("GOPATH", c.goModLayer.Root); err != nil {
		return err
	}

	c.logger.Info("Running `go install`")
	if err := c.runner.Run("go", c.appRoot, false, "install", "-buildmode", "pie", "-tags", "cloudfoundry"); err != nil {
		return err
	}

	return nil
}

func (c Contributor) getAppName() (string, error) {
	appName, err := c.runner.RunWithOutput("go", c.appRoot, false, "list", "-m")
	if err != nil {
		return "", err
	}

	return appName, nil
}

func (c Contributor) setStartCommand() error {
	appName, err := c.getAppName()
	if err != nil {
		return err
	}
	buildPath := filepath.Join(c.goModLayer.Root, "bin", appName)
	launchPath := filepath.Join(c.launchLayer.Root, appName)

	if err := c.runner.Rename(buildPath, launchPath); err != nil {
		return err
	}

	return c.launch.WriteApplicationMetadata(layers.Metadata{Processes: []layers.Process{{"web", launchPath}}})
}
