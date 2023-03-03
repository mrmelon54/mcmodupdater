package shared

type YarnVersionMeta struct {
	GameVersion string `json:"gameVersion"`
	Separator   string `json:"separator"`
	Build       int    `json:"build"`
	Maven       string `json:"maven"`
	Version     string `json:"version"`
	Stable      bool   `json:"stable"`
}

func LatestYarnVersion(v []YarnVersionMeta, mc string) (YarnVersionMeta, bool) {
	for _, i := range v {
		if i.GameVersion == mc {
			return i, true
		}
	}
	return YarnVersionMeta{}, false
}
