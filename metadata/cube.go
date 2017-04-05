package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/xiaoxiaoyijian123456/cubes/source"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
)

func Sha1Name(name string) string {
	return fmt.Sprintf("t_%s", utils.Sha1(name))
}

func NewCube() *Cube {
	return &Cube{}
}

func (c *Cube) Execute() error {
	switch c.Source.Type {
	case SOURCE_MYSQL:
		if c.Mysql == nil {
			return errors.New("ERROR: No MYSQL connection.")
		}
		sql, err := c.toSQL()
		if err != nil {
			logger.Error(err)
			return err
		}
		logger.Infof("Run cube[%s] sql: %s", c.Name, sql)
		data, err := c.Mysql.Query(sql, c.getReturnFields())
		if err != nil {
			logger.Error(err)
			return err
		}
		if err := c.saveResultToSqlite(data); err != nil {
			logger.Error(err)
			return err
		}
	case SOURCE_SQLITE:
		if c.Sqlite == nil {
			return errors.New("ERROR: No SQLITE connection.")
		}
		sql, err := c.toSQL()
		if err != nil {
			logger.Error(err)
			return err
		}
		logger.Infof("Run cube[%s] sql: %s", c.Name, sql)
		data, err := c.Sqlite.Query(sql, c.getReturnFields())
		if err != nil {
			logger.Error(err)
			return err
		}
		if err := c.saveResultToSqlite(data); err != nil {
			logger.Error(err)
			return err
		}
	case SOURCE_CUBE:
		fallthrough
	case SOURCE_JSON:
		fallthrough
	case SOURCE_CSV:
		if c.Sqlite == nil {
			return errors.New("ERROR: No sqlite connection.")
		}
		sql, err := c.toSQL()
		if err != nil {
			logger.Error(err)
			return err
		}
		logger.Infof("Run cube[%s] sql: %s", c.Name, sql)
		data, err := c.ESqlite.Query(sql, c.getReturnFields())
		if err != nil {
			logger.Error(err)
			return err
		}
		if err := c.saveResultToSqlite(data); err != nil {
			logger.Error(err)
			return err
		}
	default:
		err := errors.New(fmt.Sprintf("Unknow source type:%s", c.Source.Type))
		logger.Error(err)
		return err
	}

	return nil
}

func (c *Cube) saveResultToSqlite(data source.Rows) error {
	dmlFields := [][]string{}
	for _, v := range c.getReturnFields() {
		dmlFields = append(dmlFields, []string{v, "string"})
	}

	if err := c.ESqlite.CreateTable(c.Sha1Name, dmlFields, true); err != nil {
		logger.Error(err)
		return err
	}

	if err := c.ESqlite.InsertTable(c.Sha1Name, data); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (c *Cube) GetReport() (*CubeReport, error) {
	fields := c.getReturnFields()

	var sql bytes.Buffer
	sql.WriteString("SELECT ")
	cnt := 0
	for _, v := range fields {
		if cnt > 0 {
			sql.WriteString(fmt.Sprintf(", %s", v))
		} else {
			sql.WriteString(v)
		}
		cnt++
	}
	sql.WriteString(fmt.Sprintf(" FROM %s", c.Sha1Name))
	rows, err := c.ESqlite.Query(sql.String(), fields)
	if err != nil {
		return nil, err
	}

	return &CubeReport{
		Display: c.Display,
		Fields:  fields,
		Data:    rows,
	}, nil
}
