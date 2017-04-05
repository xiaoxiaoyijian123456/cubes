package engine

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bububa/mymysql/autorc"
	_ "github.com/bububa/mymysql/thrsafe"
	_ "github.com/mattn/go-sqlite3"
	"github.com/xiaoxiaoyijian123456/cubes/metadata"
	"github.com/xiaoxiaoyijian123456/cubes/source"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
	//"math/rand"
	"os"
	"strings"
	"time"
)

const (
	MYSQL_DEFAULT   = "default"
	DEFAULT_TMP_DIR = "/tmp/"
)

type ReportEngine struct {
	e_sqlite    *source.Sqlite
	tmpdir      string
	storesLimit *metadata.StoresLimit

	mysqls  map[string]*source.Mysql
	sqlites map[string]*source.Sqlite
}

func NewReportEngine() *ReportEngine {
	engine := &ReportEngine{
		tmpdir: getDefaultTmpDir(),

		mysqls:  make(map[string]*source.Mysql),
		sqlites: make(map[string]*source.Sqlite),
	}
	return engine
}

func (e *ReportEngine) ExecuteJsonFile(rptTplFile string, tplCfgFile string) (map[string]*metadata.CubeReport, error) {
	rptJson, err := metadata.LoadReportJsonFile(rptTplFile, tplCfgFile)
	if err != nil {
		return nil, err
	}
	return e.ExecuteRptJson(rptJson)
}

func (e *ReportEngine) ExecuteJsonConfig(jsonTpl string, tplCfg string) (map[string]*metadata.CubeReport, error) {
	cfg := make(metadata.TplCfg)
	if tplCfg != "" {
		if err := json.Unmarshal([]byte(tplCfg), &cfg); err != nil {
			logger.Errorf("ERROR Unmarshal: %v", err.Error())
			return nil, err
		}
	}
	logger.Infof("TplCfg:%v", utils.Json(cfg))

	if cfg != nil && len(cfg) > 0 {
		jsonTpl = cfg.ReplaceTpl(jsonTpl)
	}
	if strings.Contains(jsonTpl, metadata.JSON_TPL_SEP) {
		return nil, errors.New("Report Tpl still has variables.")
	}

	rptJson := metadata.ReportJson{}
	if err := json.Unmarshal([]byte(jsonTpl), &rptJson); err != nil {
		logger.Errorf("ERROR Unmarshal: %v", err.Error())
		return nil, err
	}
	return e.ExecuteRptJson(&rptJson)
}

func (e *ReportEngine) ExecuteRptJson(rptJson *metadata.ReportJson) (map[string]*metadata.CubeReport, error) {
	if err := e.initSqlite(); err != nil {
		logger.Error(err)
		return nil, err
	}
	defer e.Cleanup()

	//logger.Infof("rptJson:%s", utils.Json(rptJson))

	rpt := metadata.NewReport()
	rpt.Comment = rptJson.Comment

	if rptJson.Cube != nil {
		if err := e.ConvertCubeFromJson(rptJson.Cube, rpt); err != nil {
			logger.Error(err)
			return nil, err
		}
	}

	if rptJson.Cubes != nil {
		for _, v := range rptJson.Cubes.CubeList {
			if err := e.ConvertCubeFromJson(v, rpt); err != nil {
				logger.Error(err)
				return nil, err
			}
		}
	}

	if rptJson.CubesGroup != nil {
		for _, cubes := range rptJson.CubesGroup.CubesList {
			for _, v := range cubes.CubeList {
				if err := e.ConvertCubeFromJson(v, rpt); err != nil {
					logger.Error(err)
					return nil, err
				}
			}
		}
	}
	for _, v := range rptJson.Report {
		map_val := rpt.Cubes.Get(v)
		if map_val == nil {
			return nil, errors.New(fmt.Sprintf("Cube not found:%s", v))
		}
		cube, ok := map_val.(*metadata.Cube)
		if !ok {
			return nil, errors.New("Map data should return cube.")
		}
		rpt.Report = append(rpt.Report, cube.Name)
	}
	//logger.Info(utils.Json(rpt))
	return rpt.Execute()
}

func (e *ReportEngine) ConvertCubeFromJson(cubeJson *metadata.CubeJson, rpt *metadata.Report) error {
	cube, err := cubeJson.Convert()
	//logger.Infof("cube:%s", utils.Json(cube))
	if err != nil {
		logger.Error(err)
		return err
	}
	if v := rpt.Cubes.Get(cube.Name); v != nil {
		return errors.New(fmt.Sprintf("ERROR: duplicate cube name[%s]", cube.Name))
	}
	cube.ESqlite = e.e_sqlite

	switch cube.Source.Type {
	case metadata.SOURCE_MYSQL:
		mysql := e.getMysql(cube.Source.Name)
		if mysql == nil {
			return errors.New(fmt.Sprintf("No mysql connection:%s", cube.Source.Name))
		}
		// verify mysql connection
		mysql.Query("select 1 AS X from dual", []string{"X"})
		cube.Mysql = mysql
	case metadata.SOURCE_SQLITE:
		sqlite := e.getSqlite(cube.Source.Name)
		if sqlite == nil {
			return errors.New(fmt.Sprintf("No sqlite connection:%s", cube.Source.Name))
		}
		cube.Sqlite = sqlite
		if len(cube.Tags) > 0 {
			if err := sqlite_tags_mapping(cube.Sqlite, cube.Store.Name, cube.Tags); err != nil {
				logger.Error(err)
				return err
			}
			cube.Store.Name = fmt.Sprintf("%s_with_tags", cube.Store.Name)
			cube.Store.Sha1Name = cube.Store.Name
		}
	case metadata.SOURCE_CSV:
		cube.Sqlite = e.e_sqlite
		cube.Store = &metadata.Store{
			Name:  fmt.Sprintf("%s_csv%d", cube.Name, 1),
			Alias: fmt.Sprintf("csv%d", 1),
		}
		cube.Store.Sha1Name = metadata.Sha1Name(cube.Store.Name)
		if err := create_table_from_csv(cube.Sqlite, cube.Store.Sha1Name, cube.Source.Name, true, cube.Tags); err != nil {
			logger.Error(err)
			return err
		}
	case metadata.SOURCE_JSON:
		cube.Sqlite = e.e_sqlite
		cube.Store = &metadata.Store{
			Name:  fmt.Sprintf("%s_json%d", cube.Name, 1),
			Alias: fmt.Sprintf("json%d", 1),
		}
		cube.Store.Sha1Name = metadata.Sha1Name(cube.Store.Name)
		if err := create_table_from_json(cube.Sqlite, cube.Store.Sha1Name, cube.Source.Name, cube.Tags); err != nil {
			logger.Error(err)
			return err
		}
	case metadata.SOURCE_CUBE:
		cube.Sqlite = e.e_sqlite
	default:
		err := errors.New(fmt.Sprintf("Unknow source type:%s", cube.Source.Type))
		logger.Error(err)
		return err
	}

	cube.StoresLimit = e.storesLimit
	rpt.Cubes.Set(cube.Name, cube)
	return nil
}

func (e *ReportEngine) initSqlite() error {
	if e.e_sqlite != nil {
		e.e_sqlite.Cleanup(true)
	}

	dbname := fmt.Sprintf("%sd%d.db", e.tmpdir, time.Now().UnixNano())
	logger.Infof("sqlite dbname: %s", dbname)
	os.Remove(dbname)
	db, err := sql.Open("sqlite3", dbname)
	if err != nil {
		logger.Error(err)
		return err
	}

	e.e_sqlite = source.NewSqlite(dbname, db)
	return nil
}

func (e *ReportEngine) Cleanup() {
	if e.e_sqlite != nil {
		e.e_sqlite.Cleanup(true)
		e.e_sqlite = nil
	}

	for _, v := range e.sqlites {
		v.Cleanup(false)
	}
}

func (e *ReportEngine) AddMysqlConn(db *autorc.Conn, name string) {
	if db != nil {
		if name == "" {
			name = MYSQL_DEFAULT
		}

		e.mysqls[name] = source.NewMysql(db)
	}
}

func (e *ReportEngine) SetTmpDir(dir string) error {
	fileInfo, err := os.Stat(dir)
	if err != nil || !fileInfo.IsDir() {
		return errors.New(fmt.Sprintf("Dir[%s] not exists.", dir))
	}
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	e.tmpdir = dir
	return nil
}

func (e *ReportEngine) SetStoresLimit(storesLimit *metadata.StoresLimit) {
	e.storesLimit = storesLimit
}

func (e *ReportEngine) getMysql(dbname string) *source.Mysql {
	if _, ok := e.mysqls[MYSQL_DEFAULT]; !ok {
		mysql, err := get_default_mysql()
		if err == nil {
			e.AddMysqlConn(mysql, MYSQL_DEFAULT)
		}
	}

	if dbname == "" {
		dbname = MYSQL_DEFAULT
	}
	mysql, ok := e.mysqls[dbname]
	if !ok {
		return nil
	}

	return mysql
}

func (e *ReportEngine) getSqlite(dbname string) *source.Sqlite {
	sqlite, ok := e.sqlites[dbname]
	if ok {
		return sqlite
	}

	db, err := sql.Open("sqlite3", dbname)
	if err != nil {
		logger.Error(err)
		return nil
	}
	//defer db.Close()
	sqlite = source.NewSqlite(dbname, db)
	e.sqlites[dbname] = sqlite
	return sqlite
}

func getDefaultTmpDir() string {
	fileInfo, err := os.Stat(DEFAULT_TMP_DIR)
	if err == nil && fileInfo.IsDir() {
		return DEFAULT_TMP_DIR
	}

	return "./"
}
