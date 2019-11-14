package main

import (
	"os"
	"path"

	"github.com/NiR-/webdf/pkg/builder"
	"github.com/NiR-/webdf/pkg/pkgsolver"
	"github.com/sirupsen/logrus"
	dpkg "github.com/snyh/go-dpkg-parser"
	"github.com/spf13/cobra"
)

var updateFlags = struct {
	file string
}{}

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update",
		DisableAutoGenTag: true,
		Short:             "Update version locks",
		Run:               HandleUpdateCmd,
	}
	cmd.Flags().StringVarP(&updateFlags.file, "file", "f", "webdf.yml", "Path to the webdf.yml file to use. Webdf looks for a lock file with the same filename but with \".lock\" extension.")

	return cmd
}

func HandleUpdateCmd(cmd *cobra.Command, args []string) {
	reg := buildTypeRegistry()
	pkgSolver, err := initPackageSolver()
	if err != nil {
		logrus.Fatal(err)
	}

	b := builder.Builder{Registry: reg, PkgSolver: pkgSolver}
	if err := b.UpdateLockFile(updateFlags.file); err != nil {
		logrus.Fatal(err)
	}
}

func initPackageSolver() (pkgsolver.PackageSolver, error) {
	var pkgSolver pkgsolver.PackageSolver

	basepath := os.Getenv("XDG_DATA_HOME")
	if basepath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return pkgSolver, err
		}
		basepath = path.Join(home, ".local/share")
	}

	path := path.Join(basepath, "webdf")
	dpkgRepo := dpkg.NewRepository(path)
	pkgSolver = pkgsolver.NewPackageSolver(dpkgRepo)

	return pkgSolver, nil
}
