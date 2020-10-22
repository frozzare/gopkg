package main

import (
	"fmt"
	"github.com/go-pkg-org/gopkg/internal/archive"
	"github.com/go-pkg-org/gopkg/internal/build"
	"github.com/go-pkg-org/gopkg/internal/cache"
	"github.com/go-pkg-org/gopkg/internal/config"
	make2 "github.com/go-pkg-org/gopkg/internal/make"
	"github.com/go-pkg-org/gopkg/internal/sign"
	"github.com/go-pkg-org/gopkg/internal/upload"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).
		Level(zerolog.DebugLevel)

	app := cli.App{
		Name:    "gopkg",
		Version: "0.1.0",
		Usage:   "Reliable package manager for Golang software",
		Authors: []*cli.Author{
			{"Aloïs Micard", "alois@micard.lu"},
			{"Fredrik Forsmo", "hello@frozzare.com"},
			{"Johannes Tegnér", "johannes@jitesoft.com"},
		},
		Commands: []*cli.Command{
			{
				Name:      "make",
				Usage:     "create a new package from import-path",
				ArgsUsage: "import-path",
				Action:    execMake,
			},
			{
				Name:      "build",
				Usage:     "build a package from control directory/package",
				ArgsUsage: "control-path",
				Action:    execBuild,
			},
			{
				Name:      "install",
				Usage:     "install a package from path",
				ArgsUsage: "pkg",
				Action:    execInstall,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name: "from-file",
					},
				},
			},
			{
				Name:      "remove",
				Usage:     "remove installed package",
				ArgsUsage: "pkg-name",
				Action:    execRemove,
			},
			{
				Name:  "list",
				Usage: "list packages",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "installed",
						Usage: "list only installed packages",
					},
				},
				Action: execList,
			},
			{
				Name:   "sign",
				Usage:  "sign given package",
				Action: execSign,
			},
			{
				Name:      "upload",
				Usage:     "upload given package",
				ArgsUsage: "pkg-path",
				Action:    execUpload,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Err(err).Msg("error while running application")
		os.Exit(1)
	}
}

func execMake(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("missing import-path")
	}
	return make2.Make(c.Args().First())
}

func execBuild(c *cli.Context) error {
	path := c.Args().First()
	if path == "" {
		path = "."
	}

	absolutePath, err := getAbsolutePath(path)
	if err != nil {
		return err
	}

	return build.Build(absolutePath)
}

func execInstall(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("missing pkg")
	}

	ca, err := getCache()
	if err != nil {
		return err
	}

	if c.Bool("from-file") {
		p, err := ca.InstallPkgFile(c.Args().First())
		if err != nil {
			return fmt.Errorf("error while installing package from file %s: %s", c.Args().First(), err)
		}

		log.Info().Str("package", p.Alias).Msg("Successfully installed package")
		return nil
	}

	p, err := ca.InstallPkg(c.Args().First())
	if err != nil {
		return fmt.Errorf("error while installing package %s: %s", c.Args().First(), err)
	}
	log.Info().Str("package", p.Alias).Msg("Successfully installed package")

	return nil
}

func execRemove(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("missing pkg-name")
	}

	ca, err := getCache()
	if err != nil {
		return err
	}

	if err := ca.RemovePkg(c.Args().First()); err != nil {
		return fmt.Errorf("error while removing package %s: %s", c.Args().First(), err)
	}

	log.Info().Str("package", c.Args().First()).Msg("successfully removed package")
	return nil
}

func execList(c *cli.Context) error {
	ca, err := getCache()
	if err != nil {
		return err
	}

	pkgs, err := ca.ListPackages(c.Bool("installed"))
	if err != nil {
		return fmt.Errorf("error while listing packages: %s", err)
	}

	for _, pkg := range pkgs {
		log.Info().Str("package", pkg).Msg("")
	}

	return nil
}

func execUpload(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("missing pkg-path")
	}

	return upload.Upload(c.Args().First(), "http://127.0.0.1:8888")
}

func execSign(c *cli.Context) error {
	if !c.Args().Present() {
		return fmt.Errorf("missing pkg-path")
	}

	return sign.Sign(c.Args().First())
}

func getAbsolutePath(path string) (string, error) {
	if !filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(wd, path), nil
	}

	return path, nil
}

func getCache() (cache.Cache, error) {
	arcClient, err := archive.NewClient(archive.DefaultURL)
	if err != nil {
		return nil, err
	}

	cachePath, err := config.GetCachePath()
	if err != nil {
		return nil, err
	}

	return cache.NewCache(cachePath, arcClient)
}
