package metadata

import (
	"github.com/xiaoxiaoyijian123456/cubes/source"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
	"regexp"
)

const (
	SOURCE_MYSQL  = "mysql"
	SOURCE_SQLITE = "sqlite"
	SOURCE_CSV    = "csv"
	SOURCE_JSON   = "json"
	SOURCE_CUBE   = "cube"
)

const (
	INNER_JOIN = "INNER JOIN"
	LEFT_JOIN  = "LEFT JOIN"
	RIGHT_JOIN = "RIGHT JOIN"
)

const (
	ORDER_ASC  = "ASC"
	ORDER_DESC = "DESC"
)

const (
	SUM   = "SUM"
	AVG   = "AVG"
	COUNT = "COUNT"
	MIN   = "MIN"
	MAX   = "MAX"
)

const (
	UNION     = "UNION"
	UNION_ALL = "UNION ALL"
)

type Report struct {
	Report  []string       `json:"report,omitempty"`
	Comment string         `json:"_comment,omitempty"`
	Cubes   *utils.MapData `json:"cubes,omitempty"`
}

type Cube struct {
	Name       string                   `json:"name,omitempty"`
	Sha1Name   string                   `json:"sha1_name,omitempty"`
	Comment    string                   `json:"_comment,omitempty"`
	Display    string                   `json:"display,omitempty"` // 在结果中原样返回，用户前段显示数据用
	Source     *Source                  `json:"source,omitempty"`
	Store      *Store                   `json:"store,omitempty"`
	Join       []*Join                  `json:"join,omitempty"`
	Filter     []*AndConditions         `json:"filter,omitempty"`
	Mappings   []*Mapping               `json:"mappings,omitempty"`
	Dimensions []string                 `json:"dimensions,omitempty"`
	OrderBy    []*OrderBy               `json:"orderby,omitempty"`
	Limit      *Limit                   `json:"limit,omitempty"`
	Aggregates map[string][]*Mapping    `json:"aggregates,omitempty"`
	Tags       map[string][]*TagMapping `json:"tags,omitempty"`
	Union      []*Union                 `json:"Union,omitempty"`
	Sql        string                   `json:"sql,omitempty"`

	StoresLimit *StoresLimit `json:"stores_limit,omitempty"`

	ESqlite *source.Sqlite `json:"-"`
	Sqlite  *source.Sqlite `json:"-"`
	Mysql   *source.Mysql  `json:"-"`
}

type Source struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}
type Mapping struct {
	Alias      string `json:"alias,omitempty"`
	Expression string `json:"expression,omitempty"`
}
type TagMapping struct {
	TagVal        string         `json:"tag_val,omitempty"`
	Field         string         `json:"field,omitempty"`
	Op            string         `json:"op,omitempty"`
	Val           string         `json:"val,omitempty"`
	Val2          string         `json:"val2,omitempty"`
	IncludeRegexp *regexp.Regexp `json:"-"`
	ExcludeRegexp *regexp.Regexp `json:"-"`
}

type Condition struct {
	Field string `json:"field,omitempty"`
	Op    string `json:"op,omitempty"`
	Val   string `json:"val,omitempty"`
	Val2  string `json:"val2,omitempty"`
}

type AndConditions struct {
	Conds []*Condition `json:"conds,omitempty"`
}

type Store struct {
	Name     string `json:"name,omitempty"`
	Sha1Name string `json:"sha1_name,omitempty"`
	Alias    string `json:"alias,omitempty"`
}

type OrderBy struct {
	Field string `json:"field,omitempty"`
	Order string `json:"order,omitempty"`
}

type Limit struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

type Union struct {
	Name      string `json:"name,omitempty"`
	Sha1Name  string `json:"sha1_name,omitempty"`
	UnionType string `json:"union_type,omitempty"`
}

type Join struct {
	Type  string       `json:"type,omitempty"`
	Store *Store       `json:"store,omitempty"`
	On    []*Condition `json:"on,omitempty"`
}

type CubeReport struct {
	Display string      `json:"display"` // 在结果中原样返回，用户前段显示数据用
	Fields  []string    `json:"fields"`
	Data    source.Rows `json:"data"`
}
