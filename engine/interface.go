package engine

import (
	log "github.com/kdar/factorlog"
	"github.com/xiaoxiaoyijian123456/cubes/metadata"
	"github.com/xiaoxiaoyijian123456/cubes/source"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
)

var (
	logger *log.FactorLog = utils.SetGlobalLogger("")
)

func SetLogger(l *log.FactorLog) {
	logger = l

	metadata.SetLogger(logger)
	source.SetLogger(logger)
}
