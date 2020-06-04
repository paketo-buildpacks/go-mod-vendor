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
	Dependency = "go-mod"
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
		sort.Sort(sort.StringSlice(ldflags))
		args = append(args, fmt.Sprintf("-ldflags=%s", strings.Join(ldflags, " ")))
	}

	for _, target := range c.config.Targets {
		args = append(args, target)
	}

	c.logger.Info("Running `go install`")
	if err := c.runner.Run("go", c.appRoot, false, args...); err != nil {
		return err
	}

	binaries, err := c.getBinaryNames()
	if err != nil {
		return err
	}

	// go-install installs executables in $GOPATH/bin when not overridden by GOBIN.
	modBinPath := filepath.Join(c.goModLayer.Root, "bin")
	launchBinPath := filepath.Join(c.launchLayer.Root, "bin")
	if err := os.MkdirAll(launchBinPath, os.ModePerm); err != nil {
		return err
	}

	for _, bin := range binaries {
		binPath := filepath.Join(modBinPath, bin)
		newBinPath := filepath.Join(launchBinPath, bin)

		if _, err := os.Stat(binPath); os.IsNotExist(err) {
			c.logger.Info("`go install` failed to install executable(s) in %s",
				modBinPath)
			return err
		}

		if err := os.Rename(binPath, newBinPath); err != nil {
			return err
		}
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
	bins, err := c.getBinaryNames()
	if err != nil {
		return err
	}

	c.appName = bins[0]
	return nil
}

func (c *Contributor) getBinaryNames() ([]string, error) {
	var binaries []string
	if len(c.config.Targets) != 0 {
		for _, t := range c.config.Targets {
			appTarget := strings.TrimSuffix(t, "/")
			targetSegments := strings.Split(appTarget, "/")
			binaries = append(binaries, targetSegments[len(targetSegments) - 1])
		}
	} else if c.appName != "" {
		binaries = []string{c.appName}
	} else {
		output, err := c.runner.RunWithOutput("go", c.appRoot, false, "list", "-m")
		if err != nil {
			return []string{}, err
		}

		binaries = []string{parseAppNameFromOutput(output)}
	}

	return binaries, nil
}

func (c Contributor) setStartCommand() error {
	c.logger.Info("Contributing start command")
	launchPath := filepath.Join(c.launchLayer.Root, "bin", c.appName)

	return c.launch.WriteApplicationMetadata(layers.Metadata{
		Processes: []layers.Process{
			{
				Type:    "web",
				Command: launchPath,
				Direct:  c.context.Stack == "org.cloudfoundry.stacks.tiny" || c.context.Stack == "io.paketo.stacks.tiny",
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
