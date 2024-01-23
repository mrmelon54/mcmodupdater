package main

import (
	"flag"
	"fmt"
	"github.com/mrmelon54/mcmodupdater"
	"github.com/mrmelon54/mcmodupdater/config"
	"github.com/mrmelon54/mcmodupdater/develop"
	"github.com/mrmelon54/mcmodupdater/develop/dev"
	"io/fs"
	"os"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get CWD:", err)
		return
	}

	var dryFlag bool
	var noCache bool
	var mcVersion string
	var wdPath string

	flag.BoolVar(&dryFlag, "d", false, "Dry-run outputs the generated properties file instead of editing the file")
	flag.BoolVar(&noCache, "nocache", false, "Use flag to disable cache")
	flag.StringVar(&mcVersion, "mc", "", "Select the Minecraft version to update to, defaults to the current version")
	flag.StringVar(&wdPath, "p", cwd, "Change project path (defaults to current directory)")
	flag.Parse()

	conf, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load config")
		fmt.Println(err)
		os.Exit(1)
	}
	err = conf.Save()
	if err != nil {
		fmt.Println("Failed to save config")
		fmt.Println(err)
		os.Exit(1)
	}
	if noCache {
		conf.Cache = false
	}

	mcm, err := mcmodupdater.NewMcModUpdater(conf)
	if err != nil {
		errPrintln("Error:", err)
		os.Exit(1)
	}

	tree := os.DirFS(wdPath).(fs.StatFS)
	info, err := mcm.LoadTree(tree)
	if err != nil {
		errPrintln("Error:", err)
		os.Exit(1)
	}

	a := make([]develop.Develop, 0)
	if len(mcm.PlatArch().SubPlatforms) > 0 {
		a = append(a, mcm.PlatArch())
		for _, i := range dev.Platforms {
			if c, ok := mcm.PlatArch().SubPlatforms[i]; ok {
				a = append(a, c)
			}
		}
	}

	errPrintln("[+] Fetching version data...")

	// if the platform is architectury then fetch the sub-platform caches
	if arc, ok := info.Platform.(*dev.Architectury); ok {
		// fetch architectury specific caches first
		err := fetchCalls(arc)
		if err != nil {
			errPrintln("Error:", err)
			os.Exit(1)
		}

		// fetch sub-platform caches
		for _, i := range dev.Platforms {
			if c, ok := arc.SubPlatforms[i]; ok {
				err := fetchCalls(c)
				if err != nil {
					errPrintln("Error:", err)
					os.Exit(1)
				}
			}
		}
	} else {
		// fetch platform caches
		err := fetchCalls(info.Platform)
		if err != nil {
			errPrintln("Error:", err)
			os.Exit(1)
		}
	}

	//oldMc := info.Versions[develop.MinecraftVersion]
	info.Versions[develop.MinecraftVersion] = mcVersion
	ver := mcm.VersionUpdateList(info)

	if dryFlag {
		// output the updated properties file to stdout
		err := mcm.UpdateToVersion(os.Stdout, tree, ver.ChangeToLatest())
		if err != nil {
			errPrintln("[-] Failed to update version numbers:", err)
			os.Exit(1)
		}
	} else {
		// create temporary update file
		// this prevents accidentally destroying the original gradle properties
		uMcm, err := os.Create(".update.mcmodupdater")
		if err != nil {
			errPrintln("[-] Failed to open '.update.mcmodupdater'")
			os.Exit(1)
		}
		//goland:noinspection GoUnhandledErrorResult
		defer uMcm.Close()

		// output the updated properties file
		err = mcm.UpdateToVersion(uMcm, tree, ver.ChangeToLatest())
		if err != nil {
			errPrintln("[-] Failed to update version numbers:", err)
			os.Exit(1)
		}

		// if everything succeeded then move the temporary update file
		// to 'gradle.properties'
		err = os.Rename(".update.mcmodupdater", "gradle.properties")
		if err != nil {
			errPrintln("[-] Failed to move '.update.mcmodupdater' => '.gradle.properties'")
			os.Exit(1)
		}

		errPrintln("[+] Automatic update succeeded")
	}
}

func fetchCalls(platform develop.Develop) error {
	for _, i := range platform.FetchCalls() {
		err := i.Call()
		if err != nil {
			return err
		}
	}
	return nil
}

func errPrintln(a ...any) {
	_, _ = fmt.Fprintln(os.Stderr, a...)
}
