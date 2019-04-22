package mod

import (
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const (
	Dependency = "go-mod"
)

type Runner interface {
	Run(bin, dir string, quiet bool, args ...string) error
	RunWithOutput(bin, dir string, quiet bool, args ...string) (string, error)
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
		goModMetadata: nil,
		runner:        runner,
		appRoot:       context.Application.Root,
		logger:        context.Logger,
		launch:        context.Layers,
	}

	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	if err := c.goModLayer.Contribute(c.goModMetadata, c.contributeGoModules, c.flags()...); err != nil {
		return err
	}

	if err := c.Install(); err != nil {
		return err
	}

	return c.setStartCommand()
}

func (c Contributor) contributeGoModules(layer layers.Layer) error {
	return nil
}

func (c Contributor) flags() []layers.Flag {
	var flags []layers.Flag

	return flags
}

func (c Contributor) Install() error {
	c.logger.Info("Running `go build`")
	if err := c.runner.Run("go", c.appRoot, false, "build", "-buildmode", "pie", "-tags", "cloudfoundry"); err != nil {
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

	return c.launch.WriteApplicationMetadata(layers.Metadata{Processes: []layers.Process{{"web", "./" + appName}}})
}
