package mod

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const (
	Dependency = "go-mod-vendor"
	Cache      = "go-cache"
	Launch     = "app-binary"
)

type Runner interface {
	Run(bin, dir string, quiet bool, args ...string) error
	RunWithOutput(bin, dir string, quiet bool, args ...string) (string, error)
	SetEnv(variableName string, path string) error
}

type Logger interface {
	Info(format string, args ...interface{})
}

type Contributor struct {
	context      build.Build
	goModLayer   layers.Layer
	goCacheLayer layers.Layer
	launchLayer  layers.Layer
	runner       Runner
	appRoot      string
	logger       Logger
	launch       layers.Layers
	appName      string
	config       Config
}

func NewContributor(context build.Build, runner Runner) Contributor {
	return Contributor{
		context:      context,
		goModLayer:   context.Layers.Layer(Dependency),
		goCacheLayer: context.Layers.Layer(Cache),
		launchLayer:  context.Layers.Layer(Launch),
		runner:       runner,
		appRoot:      context.Application.Root,
		logger:       context.Logger,
		launch:       context.Layers,
	}
}

func (c Contributor) Contribute() error {
	var err error

	c.config, err = LoadConfig(c.appRoot)
	if err != nil {
		return err
	}

	c.logger.Info("Setting environment variables")
	if err := c.runner.SetEnv("GOPATH", c.goModLayer.Root); err != nil {
		return err
	}

	if err := c.runner.SetEnv("GOCACHE", c.goCacheLayer.Root); err != nil {
		return err
	}

	if err := c.goModLayer.Contribute(nil, c.ContributeGoModules, layers.Cache); err != nil {
		return err
	}

	if err := c.setAppName(); err != nil {
		return err
	}

	if err := c.launchLayer.Contribute(nil, c.ContributeBinLayer, layers.Launch); err != nil {
		return err
	}

	if err := c.goCacheLayer.Contribute(nil, c.ContributeCacheLayer, layers.Cache); err != nil {
		return err
	}

	return c.setStartCommand()
}

func (c Contributor) ContributeGoModules(_ layers.Layer) error {
	if exists, err := helper.FileExists(filepath.Join(c.appRoot, "vendor")); err != nil {
		return err
	} else if exists {
		return nil
	}
	return c.runner.Run("go", c.appRoot, false, "mod", "download")
}

func (c Contributor) ContributeBinLayer(_ layers.Layer) error {
	args := []string{"install", "-buildmode", "pie", "-tags", "cloudfoundry"}

	if exists, err := helper.FileExists(filepath.Join(c.appRoot, "vendor")); err != nil {
		return err
	} else if exists {
		args = append(args, "-mod=vendor")
	}

	if len(c.config.LDFlags) > 0 {
		var ldflags []string
		for ldflagKey, ldflagValue := range c.config.LDFlags {
			ldflags = append(ldflags, fmt.Sprintf(`-X '%s=%s'`, ldflagKey, ldflagValue))
		}
		sort.Strings(ldflags)
		args = append(args, fmt.Sprintf("-ldflags=%s", strings.Join(ldflags, " ")))
	}

	args = append(args, c.config.Targets...)

	// go-install installs executables in $GOBIN, the $GOBIN is set to the application launch layer.
	if err := os.MkdirAll(filepath.Join(c.launchLayer.Root, "bin"), os.ModePerm); err != nil {
		return err
	}

	err := c.runner.SetEnv("GOBIN", filepath.Join(c.launchLayer.Root, "bin"))
	if err != nil {
		return err
	}

	c.logger.Info("Running `go install`")
	if err := c.runner.Run("go", c.appRoot, false, args...); err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(c.launchLayer.Root, "bin", c.appName)); os.IsNotExist(err) {
		c.logger.Info("`go install` failed to install executable(s) in %s",
			filepath.Join(c.launchLayer.Root, "bin"))
		return err
	}

	return nil
}

func (c Contributor) ContributeCacheLayer(_ layers.Layer) error {
	// The cache should already exist in the layer, we do this to write empty metadata so that this layer will always be retrieved
	return nil
}

func (c Contributor) Cleanup() error {
	contents, err := filepath.Glob(filepath.Join(c.appRoot, "*"))
	if err != nil {
		return err
	}

	for _, file := range contents {
		if err := os.RemoveAll(file); err != nil {
			return err
		}
	}

	return nil
}

func (c *Contributor) setAppName() error {
	if len(c.config.Targets) != 0 {
		appTarget := strings.TrimSuffix(c.config.Targets[0], "/")
		targetSegments := strings.Split(appTarget, "/")
		appName := targetSegments[len(targetSegments)-1]
		c.appName = appName
	} else {
		output, err := c.runner.RunWithOutput("go", c.appRoot, false, "list", "-m")
		if err != nil {
			return err
		}

		c.appName = parseAppNameFromOutput(output)
	}

	return nil
}

func (c Contributor) setStartCommand() error {
	c.logger.Info("Contributing start command")
	launchPath := filepath.Join(c.launchLayer.Root, "bin", c.appName)

	return c.launch.WriteApplicationMetadata(layers.Metadata{
		Processes: []layers.Process{
			{
				Type:    "web",
				Command: launchPath,
				Direct:  c.context.Stack == "io.paketo.stacks.tiny",
			},
		},
	})
}

func parseAppNameFromOutput(output string) string {
	sanitizedOutput := sanitizeOutput(output)
	moduleNamePaths := strings.Split(sanitizedOutput, "/")
	return moduleNamePaths[len(moduleNamePaths)-1]
}

func sanitizeOutput(output string) string {
	lines := strings.Split(output, "\n")
	return lines[len(lines)-1]
}
