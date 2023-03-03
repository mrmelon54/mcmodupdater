package develop

type PropVersion int

func (v PropVersion) Key() string { return propVersionKeyMap[v] }

// If this doesn't work then go generate hasn't been run
var _ = PropVersion(0).String()

func PropVersionFromKey(k string) (PropVersion, bool) {
	a, ok := propVersionFromKeys[k]
	return a, ok
}

//go:generate stringer -type=PropVersion -linecomment

const (
	_                     = PropVersion(iota)
	ModVersion            // Version
	MinecraftVersion      // Minecraft
	ArchitecturyVersion   // Architectury
	FabricLoaderVersion   // Fabric Loader
	FabricApiVersion      // Fabric API
	YarnMappingsVersion   // Yarn Mappings
	ForgeVersion          // Forge
	ForgeMappingsVersion  // Forge Mappings
	QuiltLoaderVersion    // Quilt Loader
	QuiltFabricApiVersion // Quilted Fabric API
	QuiltMappingsVersion  // Quilt Mappings
)

var (
	propVersionKeyMap = map[PropVersion]string{
		ModVersion:            "mod_version",
		MinecraftVersion:      "minecraft_version",
		ArchitecturyVersion:   "architectury_version",
		FabricLoaderVersion:   "fabric_loader_version",
		FabricApiVersion:      "fabric_api_version",
		YarnMappingsVersion:   "yarn_mappings",
		ForgeVersion:          "forge_version",
		ForgeMappingsVersion:  "forge_mappings_version",
		QuiltLoaderVersion:    "quilt_loader_version",
		QuiltFabricApiVersion: "quilt_fabric_api_version",
		QuiltMappingsVersion:  "quilt_mappings",
	}
	// basically inverted propVersionKeyMap
	propVersionFromKeys map[string]PropVersion
)

func init() {
	propVersionFromKeys = make(map[string]PropVersion)
	for k, v := range propVersionKeyMap {
		propVersionFromKeys[v] = k
	}
}
