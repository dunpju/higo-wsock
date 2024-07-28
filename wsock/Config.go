package wsock

import "sync"

var requireUpgrade *UpgradeConfigure

func init() {
	requireUpgrade = newUpgradeConfigure()
}

type UpgradeConfigure struct {
	config *sync.Map
}

func (this *UpgradeConfigure) Config() *sync.Map {
	return this.config
}

func (this *UpgradeConfigure) Load(httpMethod, relativePath string) (value interface{}, ok bool) {
	return this.config.Load(httpMethod + "@" + relativePath)
}

func newUpgradeConfigure() *UpgradeConfigure {
	return &UpgradeConfigure{config: &sync.Map{}}
}

func UpgradeConn(httpMethod, relativePath string) {
	requireUpgrade.config.Store(httpMethod+"@"+relativePath, true)
}

func UpgradeConfig() *UpgradeConfigure {
	return requireUpgrade
}
