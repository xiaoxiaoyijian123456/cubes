package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/kdar/factorlog"
	"github.com/xiaoxiaoyijian123456/cubes/engine"
	"github.com/xiaoxiaoyijian123456/cubes/metadata"
	"github.com/xiaoxiaoyijian123456/cubes/rest"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
	"io/ioutil"
	"math/rand"
	"path"
	"runtime"
	"time"
)

const (
	MODE_REPOTS = "reports"
	MODE_SERVER = "server"
	MODE_IMPORT = "import"
	MODE_EXPORT = "export"
)

var (
	logFlag  = flag.String("log", "", "set log path")
	portFlag = flag.Int("port", 9100, "set port")
	modeFlag = flag.String("mode", "", "set running mode")

	jsonTplFlag       = flag.String("jtpl", "", "set reports json tpl file")
	jsonTplConfigFlag = flag.String("jcfg", "", "set tpl config file")
	outputFlag        = flag.String("output", "", "set reports json output")

	fileFlag  = flag.String("file", "", "import/export csv/json file name")
	dbFlag    = flag.String("db", "", "import/export database name")
	tableFlag = flag.String("table", "", "import/export table name")

	logger *log.FactorLog
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	logger = utils.SetGlobalLogger(*logFlag)
	engine.SetLogger(logger)

	mode := *modeFlag
	if mode == "" {
		mode = MODE_REPOTS
	}

	switch mode {
	case MODE_REPOTS:
		if *jsonTplFlag == "" {
			logger.Error("No json tpl file.")
			return
		}
		rptEngine := engine.NewReportEngine()
		defer rptEngine.Cleanup()

		storesLimit, err := metadata.NewStoresLimitFromJson(SkylineStoresLimitJson)
		if err != nil {
			logger.Error(err)
			return
		}
		storesLimit.SetFieldSetting("client_id", 1)
		rptEngine.SetStoresLimit(storesLimit)

		rptsResult, err := rptEngine.ExecuteJsonFile(*jsonTplFlag, *jsonTplConfigFlag)
		if err != nil {
			logger.Error(err)
			return
		}

		logger.Infof("reports result:%v", utils.Json(rptsResult))
		if *outputFlag != "" {
			if err := ioutil.WriteFile(*outputFlag, []byte(utils.Json(rptsResult)), 0666); err != nil {
				logger.Error(err)
				return
			}
		}

	case MODE_SERVER:
		rest.SetLogger(logger)
		gin.SetMode(gin.ReleaseMode)
		gin.DisableBindValidation()
		router := gin.New()
		router.Use(gin.Recovery())
		cubesGroup := router.Group("/")
		{
			cubesGroup.POST("rpts", rest.CubesRptHandler)
		}
		logger.Infof("Cubes Server started at:0.0.0.0:%d", *portFlag)
		defer func() {
			logger.Infof("Cubes Server exit from:0.0.0.0:%d", *portFlag)
		}()
		router.Run(fmt.Sprintf(":%d", *portFlag))

	case MODE_IMPORT:
		if *dbFlag == "" || *tableFlag == "" {
			logger.Error("import file: `import-db`, `imort-table` also needed")
			return
		}

		engine := engine.NewImportEngine()
		ext := path.Ext(*fileFlag)
		switch ext {
		case ".csv":
			if err := engine.ImportCsvFile(*fileFlag, *dbFlag, *tableFlag); err != nil {
				logger.Error(err)
				return
			}
		case ".json":
			if err := engine.ImportJsonFile(*fileFlag, *dbFlag, *tableFlag); err != nil {
				logger.Error(err)
				return
			}
		default:
			logger.Errorf("import file: unknown file ext: %s", ext)
		}
		return

	case MODE_EXPORT:
		if *dbFlag == "" {
			logger.Error("export table: `export-db` also needed")
			return
		}

		engine := engine.NewExportEngine()
		data, err := engine.ExportTable(*dbFlag, *tableFlag)
		if err != nil {
			logger.Error(err)
			return
		}
		exportFile := *fileFlag
		if exportFile == "" {
			exportFile = fmt.Sprintf("./%s.json", *tableFlag)
		}
		logger.Infof("exportFile = %s", exportFile)
		if err := ioutil.WriteFile(exportFile, []byte(utils.Json(data)), 0666); err != nil {
			logger.Error(err)
			return
		}
		return
	}
}

var SkylineStoresLimitJson = `
{
    "limit_stores": {
        "skyline.simba_adgroup_rpt_daily": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_campaign_rpt_daily": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_client_rpt_daily": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_keyword_rpt_daily": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_adgroup_rpt_rt": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_campaign_rpt_rt": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_client_rpt_rt": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_keyword_rpt_rt": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_adgroups": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_campaigns": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_creatives": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_creatives_store": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_items": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.simba_keywords": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.zhizuan_campaign_rpt_daily": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.zhizuan_client_rpt_daily": {
            "fields": [
                "client_id"
            ]
        },
        "skyline.zhizuan_target_rpt_daily": {
            "fields": [
                "client_id"
            ]
        }
    },
    "fields_setting": {}
}
`
