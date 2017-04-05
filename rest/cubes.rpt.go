package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/xiaoxiaoyijian123456/cubes/engine"
	"github.com/xiaoxiaoyijian123456/cubes/metadata"
	"net/http"
)

type CubesRptRequest struct {
	JsonTpl string `form:"json_tpl" binding:"required"`
	TplCfg  string `form:"tpl_cfg"`
}

func CubesRptHandler(c *gin.Context) {
	var request CubesRptRequest
	if err := c.BindWith(&request, binding.FormMultipart); err != nil {
		c.JSON(http.StatusOK, APIError{Code: BADREQUEST_ERROR, Msg: err.Error()})
		return
	}
	//logger.Infof("request = %v", Json(request))

	engine.SetLogger(logger)
	rptEngine := engine.NewReportEngine()
	defer rptEngine.Cleanup()

	storesLimit, err := metadata.NewStoresLimitFromJson(StoresLimitJson)
	if err != nil {
		c.JSON(http.StatusOK, APIError{Code: BADREQUEST_ERROR, Msg: err.Error()})
		return
	}
	rptEngine.SetStoresLimit(storesLimit)

	rptResult, err := rptEngine.ExecuteJsonConfig(request.JsonTpl, request.TplCfg)
	if err != nil {
		c.JSON(http.StatusOK, APIError{Code: BADREQUEST_ERROR, Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, rptResult)
}
