package source

import (
	log "github.com/kdar/factorlog"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
)

var (
	logger *log.FactorLog = utils.SetGlobalLogger("")
)

func SetLogger(l *log.FactorLog) {
	logger = l
}
