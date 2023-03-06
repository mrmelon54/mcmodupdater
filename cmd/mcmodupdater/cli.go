package main

import (
	"flag"
	"fmt"
	"github.com/MrMelon54/mcmodupdater"
	"github.com/MrMelon54/mcmodupdater/config"
	"github.com/MrMelon54/mcmodupdater/develop"
	"github.com/MrMelon54/mcmodupdater/develop/dev"
	"io/fs"
	"os"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Failed to get CWD:", err)
		return
	}

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

	var dryFlag bool
	var mcVersion string

	flag.BoolVar(&dryFlag, "dry", false, "Dry-run outputs the generated properties file instead of editing the file")
	flag.StringVar(&mcVersion, "mc", "", "Select the Minecraft version to update to, defaults to the current version")
	flag.Parse()

	mcm, err := mcmodupdater.NewMcModUpdater(conf)
	if err != nil {
		errPrintln("Error:", err)
		os.Exit(1)
	}

	info, err := mcm.LoadTree(os.DirFS(cwd).(fs.StatFS))
	if err != nil {
		errPrintln("Error:", err)
		os.Exit(1)
	}

	isClean, err := mcm.IsClean()
	if err != nil {
		errPrintln("[-] Failed to detect if the current worktree is clean:", err)
		os.Exit(1)
	}

	if !isClean {
		errPrintln("[-] The current worktree is dirty")
		os.Exit(1)
	}

	branch, err := mcm.CurrentBranch()
	if err != nil {
		errPrintln("[-] Failed to find current branch:", err)
		os.Exit(1)
	}

	info, err := mcm.LoadBranch(branch)
	if err != nil {
		errPrintln("[-] Failed to load branch info:", err)
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

	ver := mcm.VersionUpdateList(info)

	if dryFlag {
		// output the updated properties file to stdout
		err := mcm.UpdateToVersion(os.Stdout, ver.ChangeToLatest())
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
		err = mcm.UpdateToVersion(uMcm, ver.ChangeToLatest())
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
