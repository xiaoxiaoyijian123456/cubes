package metadata

import (
	"encoding/json"
	"fmt"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
	"io/ioutil"
	"strings"
)

const JSON_TPL_SEP = "#####"

type TplCfg map[string]interface{}

func (t TplCfg) ReplaceTpl(reportJson string) string {
	for k, v := range t {
		reportJson = strings.Replace(reportJson, fmt.Sprintf("%s%s%s", JSON_TPL_SEP, utils.UpperTrim(k), JSON_TPL_SEP), fmt.Sprintf("%v", v), -1)
	}

	return reportJson
}

func LoadTplCfgFile(tplCfgFile string) (TplCfg, error) {
	if tplCfgFile == "" {
		return nil, nil
	}

	bytes, err := ioutil.ReadFile(tplCfgFile)
	if err != nil {
		logger.Errorf("ERROR: failed to read file[%s] :%v", tplCfgFile, err.Error())
		return nil, err
	}

	cfg := make(TplCfg)
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		logger.Errorf("ERROR Unmarshal: %v", err.Error())
		return nil, err
	}
	logger.Infof("TplCfg:%v", utils.Json(cfg))
	return cfg, nil
}
