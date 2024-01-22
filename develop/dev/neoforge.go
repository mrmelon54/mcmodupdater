package dev

import (
	"encoding/xml"
	"fmt"
	"github.com/MrMelon54/mcmodupdater/config"
	"github.com/MrMelon54/mcmodupdater/develop"
	"github.com/MrMelon54/mcmodupdater/meta"
	"github.com/MrMelon54/mcmodupdater/meta/shared"
	"github.com/magiconair/properties"
	"io"
	"io/fs"
	"path"
)

var (
	PlatformNeoForge        = develop.DevPlatform{Name: "NeoForge", Sub: "neoforge"}
	neoForgeLoaderMetaPaths = []string{
		"src/main/resources/META-INF/mods.toml",
		"resources/META-INF/mods.toml",
	}
)

type NeoForge struct {
	Conf  config.NeoForgeDevelopConfig
	Meta  *NeoForgeMeta
	Cache string
}

func ForNeoForge(conf config.DevelopConfig, cache string) develop.Develop {
	return &NeoForge{
		Conf:  conf.NeoForge,
		Meta:  &NeoForgeMeta{},
		Cache: path.Join(cache, "neoforge"),
	}
}

type NeoForgeMeta struct {
	done chan struct{}
	Api  meta.NeoForgeApiMeta
}

func (f *NeoForge) Platform() develop.DevPlatform {
	return PlatformNeoForge
}

func (f *NeoForge) FetchCalls() []develop.DevFetch {
	return []develop.DevFetch{
		{"API", f.fetchApi},
	}
}

func (f *NeoForge) ValidTree(tree fs.FS) bool {
	_, ok := genericCheckOnePathExists(tree, neoForgeLoaderMetaPaths...)
	return ok
}

func (f *NeoForge) ReadVersionFile(tree fs.FS) (map[develop.PropVersion]string, error) {
	gradlePropFile, err := tree.Open("gradle.properties")
	if err != nil {
		return nil, fmt.Errorf("open gradle.properties: %w", err)
	}
	gradlePropContent, err := io.ReadAll(gradlePropFile)
	if err != nil {
		return nil, fmt.Errorf("read gradle.properties: %w", err)
	}
	prop, err := properties.Load(gradlePropContent, 0)
	if err != nil {
		return nil, err
	}

	propM := prop.Map()
	a := make(map[develop.PropVersion]string)
	mapProp(a, develop.ModVersion, propM)
	mapProp(a, develop.MinecraftVersion, propM)
	mapProp(a, develop.NeoForgeVersion, propM)
	return a, nil
}

func (f *NeoForge) LatestVersion(prop develop.PropVersion, mcVersion string) (string, bool) {
	println("neoforge latest version", prop)
	switch prop {
	case develop.NeoForgeVersion:
		if a, ok := shared.LatestNeoForgeMavenVersion(shared.MavenMeta(f.Meta.Api), mcVersion); ok {
			return a, true
		}
	default:
	}
	return "", false
}

func (f *NeoForge) fetchApi() (err error) {
	f.Meta.Api, err = genericPlatformFetch[meta.NeoForgeApiMeta](f.Conf.Api, path.Join(f.Cache, "api.xml"), func(r io.Reader, m *meta.NeoForgeApiMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.NeoForgeApiMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}