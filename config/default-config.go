package config

import (
	"encoding/json"
	"github.com/mrmelon54/mcmodupdater/paths"
	"os"
	"path"
)

func DefaultConfig() Config {
	return Config{
		Develop: DevelopConfig{
			Architectury: ArchitecturyDevelopConfig{
				Api: "https://api.modrinth.com/v2/project/architectury-api/version",
			},
			Fabric: FabricDevelopConfig{
				Game:   "https://meta.fabricmc.net/v2/versions/game",
				Yarn:   "https://meta.fabricmc.net/v2/versions/yarn",
				Loader: "https://meta.fabricmc.net/v2/versions/loader",
				Api:    "https://maven.fabricmc.net/net/fabricmc/fabric-api/fabric-api/maven-metadata.xml",
			},
			Forge: ForgeDevelopConfig{
				Api: "https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml",
			},
			Quilt: QuiltDevelopConfig{
				Game:                 "https://meta.quiltmc.org/v3/versions/game",
				QuiltMappings:        "https://meta.quiltmc.org/v3/versions/quilt-mappings",
				QuiltMappingsOnLoom:  "https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-mappings-on-loom/maven-metadata.xml",
				Loader:               "https://meta.quiltmc.org/v3/versions/loader",
				QuiltStandardLibrary: "https://maven.quiltmc.org/repository/release/org/quiltmc/qsl/maven-metadata.xml",
				QuiltedFabricApi:     "https://maven.quiltmc.org/repository/release/org/quiltmc/quilted-fabric-api/quilted-fabric-api/maven-metadata.xml",
			},
			NeoForge: NeoForgeDevelopConfig{
				Api: "https://maven.neoforged.net/net/neoforged/neoforge/maven-metadata.xml",
			},
		},
		Cache: true,
	}
}

func Load() (*Config, error) {
	conf := DefaultConfig()
	vFile := path.Join(paths.UserConfigDir(), "config.json")
	f, err := os.Open(vFile)
	if err != nil {
		return &conf, nil
	}

	err = json.NewDecoder(f).Decode(&conf)
	return &conf, err
}

func (c *Config) Save() error {
	vFile := path.Join(paths.UserConfigDir(), "config.json")
	err := os.MkdirAll(path.Dir(vFile), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(vFile)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(c)
}
