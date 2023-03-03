package lib_version

type LibVersion struct {
	Versions  map[string]string   `json:"versions,omitempty"`
	Libraries map[string]Library  `json:"libraries,omitempty"`
	Bundles   map[string][]string `json:"bundles,omitempty"`
	Plugins   map[string]Plugin   `json:"plugins,omitempty"`
}

type Library struct {
	Module  string  `json:"module,omitempty"`
	Version Version `json:"version,omitempty"`
}

type Version struct {
	Ref string `json:"ref,omitempty"`
}

type Plugin struct {
	ID      string `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
}
