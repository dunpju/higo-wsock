package wsock

import "sync"

var requireUpgrade *UpgradeConfigurer

func init() {
	requireUpgrade = NewUpgradeConfigurer()
}

type UpgradeConfigurer struct {
	config *sync.Map
}

func (this *UpgradeConfigurer) Config() *sync.Map {
	return this.config
}

func (this *UpgradeConfigurer) Load(httpMethod, relativePath string) (value interface{}, ok bool) {
	return this.config.Load(httpMethod + "@" + relativePath)
}

func NewUpgradeConfigurer() *UpgradeConfigurer {
	return &UpgradeConfigurer{config: &sync.Map{}}
}

func UpgradeConn(httpMethod, relativePath string) {
	requireUpgrade.config.Store(httpMethod+"@"+relativePath, true)
}
