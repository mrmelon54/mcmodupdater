package paths

import "github.com/wessie/appdirs"

var appDirs = appdirs.New("mcmodupdater", "MrMelon54", "")

func UserConfigDir() string {
	return appDirs.UserConfig()
}

func UserCacheDir() string {
	return appDirs.UserCache()
}
