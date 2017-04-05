package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
	"strings"
)

func (c *Cube) toSQL() (string, error) {
	if len(c.Union) > 0 && c.Source.Type == SOURCE_CUBE {
		return c.toUnionSQL()
	}
	if c.Sql != "" {
		return c.Sql, nil
	}

	var buffer bytes.Buffer
	buffer.WriteString(`SELECT `)

	cnt := 0
	if len(c.Dimensions) > 0 {
		for _, v := range c.Dimensions {
			if cnt > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(v)
			cnt++
		}

		if len(c.Aggregates) > 0 {
			buffer.WriteString(", ")
		}
	}

	cnt = 0
	for f, fields := range c.Aggregates {
		for _, v := range fields {
			if cnt > 0 {
				buffer.WriteString(", ")
			}
			switch f {
			case "COUNT":
				if v.Expression == "*" || v.Expression == "1" {
					buffer.WriteString(fmt.Sprintf(" IFNULL(COUNT(*),0) AS %s", v.Alias))
				} else {
					buffer.WriteString(fmt.Sprintf(" IFNULL(COUNT(DISTINCT IFNULL(%s, 0)),0) AS %s", f, v.Expression, v.Alias))
				}
			default:
				buffer.WriteString(fmt.Sprintf(" IFNULL(%s(IFNULL(%s, 0)),0) AS %s", f, v.Expression, v.Alias))
			}
			cnt++
		}
	}
	if len(c.Dimensions) == 0 && len(c.Aggregates) == 0 {
		buffer.WriteString(fmt.Sprintf(" %s.* ", TMP_TABLE_ALIAS))
	}
	buffer.WriteString(` FROM ( `)
	buffer.WriteString(` SELECT `)
	cnt = 0
	base_select_fields := c.getDefaultStoreFields()
	base_select_fields_map := make(map[string]struct{})
	for _, v := range base_select_fields {
		base_select_fields_map[utils.LowerTrim(v)] = struct{}{}
	}

	for _, v := range c.Mappings {
		if cnt > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf("%s AS %s", v.Expression, v.Alias))
		if _, ok := base_select_fields_map[utils.LowerTrim(v.Alias)]; ok {
			delete(base_select_fields_map, utils.LowerTrim(v.Alias))
		}

		cnt++
	}
	if c.Source.Type == "mysql" && len(c.Tags) > 0 {
		cnt = 0
		for tagName, mappings := range c.Tags {
			if cnt > 0 || len(c.Mappings) > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(" CASE ")
			for _, m := range mappings {
				buffer.WriteString(fmt.Sprintf(` WHEN %s REGEXP "%s" `, m.Field, m.Val))
				if m.Val2 != "" {
					buffer.WriteString(fmt.Sprintf(` AND %s NOT REGEXP "%s" `, m.Field, m.Val2))
				}
				buffer.WriteString(fmt.Sprintf(` THEN "%s"`, m.TagVal))
			}
			buffer.WriteString(fmt.Sprintf(` ELSE "" END AS %s`, tagName))

			if _, ok := base_select_fields_map[utils.LowerTrim(tagName)]; ok {
				delete(base_select_fields_map, utils.LowerTrim(tagName))
			}

			cnt++
		}
	}
	select_fields := []string{}
	for _, v := range base_select_fields {
		if _, ok := base_select_fields_map[utils.LowerTrim(v)]; ok {
			select_fields = append(select_fields, v)
		}
	}
	if len(c.Mappings) > 0 || (len(c.Tags) > 0 && c.Source.Type == "mysql") {
		buffer.WriteString(", ")
	}
	cnt = 0
	for _, v := range select_fields {
		if cnt > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf(" %s.%s ", c.Store.Alias, v))
		cnt++
	}
	buffer.WriteString(` FROM `)
	buffer.WriteString(fmt.Sprintf("%s AS %s", c.Store.Sha1Name, c.Store.Alias))

	if len(c.Join) > 0 {
		for _, join := range c.Join {
			buffer.WriteString(fmt.Sprintf(` %s %s AS %s `, join.Type, join.Store.Sha1Name, join.Store.Alias))
			buffer.WriteString(" ON (")
			cnt := 0
			for _, cond := range join.On {
				if cnt > 0 {
					buffer.WriteString(" AND ")
				}
				cond_val := fmt.Sprintf(`"%s"`, cond.Val)
				if len(strings.Split(cond.Val, ".")) > 1 {
					cond_val = cond.Val
				}
				switch cond.Op {
				case "BETWEEN":
					buffer.WriteString(fmt.Sprintf(`%s BETWEEN "%s" AND "%s"`, cond.Field, cond.Val, cond.Val2))
				default:
					buffer.WriteString(fmt.Sprintf(`%s %s %s`, cond.Field, cond.Op, cond_val))
				}
				cnt++
			}
			if c.Source.Type == SOURCE_MYSQL && c.StoresLimit != nil && len(c.StoresLimit.FieldsSetting) > 0 {
				limitStore := c.StoresLimit.GetLimitStore(join.Store.Name)
				if limitStore != nil {
					for _, field := range limitStore.Fields {
						if v := c.StoresLimit.GetFieldSetting(field); v != nil {
							buffer.WriteString(fmt.Sprintf(" AND %s.%s = %v ", join.Store.Alias, field, v))
						}
					}
				}
			}

			buffer.WriteString(") ")
		}
	}

	buffer.WriteString(` WHERE 1=1 `)
	if c.Source.Type == SOURCE_MYSQL && c.StoresLimit != nil && len(c.StoresLimit.FieldsSetting) > 0 {
		limitStore := c.StoresLimit.GetLimitStore(c.Store.Name)
		if limitStore != nil {
			buffer.WriteString(` AND ( `)
			andCnt := 0
			for _, field := range limitStore.Fields {
				if v := c.StoresLimit.GetFieldSetting(field); v != nil {
					if andCnt > 0 {
						buffer.WriteString(" AND ")
					}
					buffer.WriteString(fmt.Sprintf(" %s.%s = %v ", c.Store.Alias, field, v))
					andCnt++
				}
			}
			buffer.WriteString(` ) `)
		}
	}
	if len(c.Filter) > 0 {
		buffer.WriteString(` AND ( `)

		cnt = 0
		for _, andConds := range c.Filter {
			if cnt > 0 {
				buffer.WriteString(" OR ")
			}
			buffer.WriteString("(")
			andCnt := 0
			for _, cond := range andConds.Conds {
				if andCnt > 0 {
					buffer.WriteString(" AND ")
				}
				cond_val := fmt.Sprintf(`"%s"`, cond.Val)
				if len(strings.Split(cond.Val, ".")) > 1 {
					cond_val = cond.Val
				}
				switch cond.Op {
				case "BETWEEN":
					buffer.WriteString(fmt.Sprintf(`%s BETWEEN "%s" AND "%s"`, cond.Field, cond.Val, cond.Val2))
				default:
					if cond_val == "" {
						buffer.WriteString(fmt.Sprintf(`%s %s ""`, cond.Field, cond.Op))
					} else {
						buffer.WriteString(fmt.Sprintf(`%s %s %s`, cond.Field, cond.Op, cond_val))
					}
				}

				andCnt++
			}
			buffer.WriteString(")")
			cnt += 1
		}

		buffer.WriteString(` ) `)
	}
	buffer.WriteString(fmt.Sprintf(`) AS %s `, TMP_TABLE_ALIAS))

	if len(c.Aggregates) > 0 && len(c.Dimensions) > 0 {
		buffer.WriteString(` GROUP BY `)
		cnt = 0
		for _, v := range c.Dimensions {
			if cnt > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(v)
			cnt++
		}
	}

	if len(c.OrderBy) > 0 {
		buffer.WriteString(` ORDER BY `)
		cnt = 0
		for _, v := range c.OrderBy {
			if cnt > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteString(fmt.Sprintf("%s %s", v.Field, v.Order))
			cnt++
		}
	}
	if c.Limit != nil {
		buffer.WriteString(fmt.Sprintf(` LIMIT %d OFFSET %d `, c.Limit.Limit, c.Limit.Offset))
	}
	return buffer.String(), nil
}

const TMP_TABLE_ALIAS = "XXXXX_TMP"

func (c *Cube) toUnionSQL() (string, error) {
	var buffer bytes.Buffer
	if len(c.Union) == 0 || c.Source.Type != SOURCE_CUBE {
		return "", errors.New("Only support cube union.")
	}

	fieldsCnt := 0
	cnt := 0
	for _, union := range c.Union {
		if cnt > 0 {
			buffer.WriteString(fmt.Sprintf(" %s ", union.UnionType))
		}

		buffer.WriteString(` SELECT `)
		fields, err := c.Sqlite.GetTableFields(union.Sha1Name)
		if err != nil {
			logger.Error(err)
			return "", err
		}
		if fieldsCnt == 0 {
			fieldsCnt = len(fields)
		}
		if len(fields) != fieldsCnt {
			return "", errors.New(fmt.Sprintf("Cube union: %s, fields cnt not same.", union.Name))
		}

		buffer.WriteString(fmt.Sprintf(`%s `, strings.Join(fields, ",")))
		buffer.WriteString(fmt.Sprintf(" FROM %s ", union.Sha1Name))

		cnt++
	}
	return buffer.String(), nil
}

func (c *Cube) getReturnFields() []string {
	if len(c.Union) > 0 && c.Source.Type == SOURCE_CUBE {
		fields, err := c.Sqlite.GetTableFields(c.Union[0].Sha1Name)
		if err != nil {
			logger.Error(err)
			return []string{}
		}
		return fields
	}

	fields := []string{}
	for _, v := range c.Dimensions {
		fields = append(fields, v)
	}
	for _, aggregate := range c.Aggregates {
		for _, v := range aggregate {
			fields = append(fields, v.Alias)
		}
	}

	if len(fields) == 0 {
		switch c.Source.Type {
		case SOURCE_CUBE:
			fallthrough
		case SOURCE_CSV:
			fallthrough
		case SOURCE_JSON:
			if retFields, err := c.Sqlite.GetTableFields(c.Store.Sha1Name); err == nil {
				fields = retFields
			}
		case SOURCE_SQLITE:
			if retFields, err := c.Sqlite.GetTableFields(c.Store.Name); err == nil {
				fields = retFields
			}
		case SOURCE_MYSQL:
			if retFields, err := c.Mysql.GetTableFields(c.Store.Name); err == nil {
				fields = retFields
			}
		}
	}
	return fields
}

func (c *Cube) getDefaultStoreFields() []string {
	var (
		fields []string
		err    error
	)
	switch c.Source.Type {
	case SOURCE_MYSQL:
		fields, err = c.Mysql.GetTableFields(c.Store.Name)
	case SOURCE_CUBE:
		fallthrough
	case SOURCE_CSV:
		fallthrough
	case SOURCE_JSON:
		fields, err = c.Sqlite.GetTableFields(c.Store.Sha1Name)
	case SOURCE_SQLITE:
		fields, err = c.Sqlite.GetTableFields(c.Store.Name)
	}
	if err != nil || len(fields) == 0 {
		logger.Error(err)
		return nil
	}
	return fields
}
