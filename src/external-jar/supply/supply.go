package supply

import (
	"io"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Stager interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/stager.go
	BuildDir() string
	DepDir() string
	DepsIdx() string
	DepsDir() string
	WriteConfigYml(interface{}) error
}

type Manifest interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/manifest.go
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
}

type Installer interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/installer.go
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type Command interface {
	//TODO: See more options at https://github.com/cloudfoundry/libbuildpack/blob/master/command.go
	Execute(string, io.Writer, io.Writer, string, ...string) error
	Output(dir string, program string, args ...string) (string, error)
}

type Supplier struct {
	Manifest  Manifest
	Installer Installer
	Stager    Stager
	Command   Command
	Log       *libbuildpack.Logger
}

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying external-jar")

	dep, err := s.Manifest.DefaultVersion("external-jar")
	if err != nil {
		return err
	}

	s.Log.Info("Using external-jar version %s", dep.Version)

	if err := s.Installer.InstallDependency(dep, s.Stager.DepDir()); err != nil {
		return err
	}

	s.Log.Info("Using extension_directories %s", dep.Version)

	// The extension directories filepath will be prefixed with "$PWD/../.." and the value of the env variable PWD will be "/home/vcap/app"
	extension_directories_filepath := filepath.Join("/" + os.Getenv("USER"), "/deps", "/" + s.Stager.DepsIdx(), "/external-jar")

	// See https://github.com/cloudfoundry/java-buildpack/blob/63545391234676b91642b7e0c5f946113ac8b3b4/docs/framework-multi_buildpack.md
	config := map[string]interface{} {
		"extension_directories" : []string {extension_directories_filepath},
	}

	if err := s.Stager.WriteConfigYml(config); err != nil {
		s.Log.Error("Error writing config.yml: %s", err.Error())
		return err
	}

	return nil
}
