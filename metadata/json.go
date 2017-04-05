package metadata

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

func LoadReportJsonFile(jsonTplFile string, tplCfgFile string) (*ReportJson, error) {
	bytes, err := ioutil.ReadFile(jsonTplFile)
	if err != nil {
		logger.Errorf("ERROR: failed to read file[%s] :%v", jsonTplFile, err.Error())
		return nil, err
	}
	jsonContent := string(bytes)

	tplCfg, err := LoadTplCfgFile(tplCfgFile)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	if tplCfg != nil && len(tplCfg) > 0 {
		jsonContent = tplCfg.ReplaceTpl(jsonContent)
	}
	if strings.Contains(jsonContent, JSON_TPL_SEP) {
		return nil, errors.New("Report Tpl still has variables.")
	}

	report := ReportJson{}
	if err := json.Unmarshal([]byte(jsonContent), &report); err != nil {
		logger.Errorf("ERROR Unmarshal: %v", err.Error())
		return nil, err
	}
	logger.Infof("ReportJson:%v", utils.Json(report))
	return &report, nil
}

func (r *CubeJson) Convert() (*Cube, error) {
	// Trim all string fields
	r.Name = utils.Trim(r.Name)
	r.Store = utils.Trim(r.Store)
	r.Dimensions = utils.Trim(r.Dimensions)
	r.Limit = utils.Trim(r.Limit)
	r.Source = utils.Trim(r.Source)

	cube := &Cube{
		Name:       r.Name,
		Comment:    utils.Trim(r.Comment),
		Display:    utils.Trim(r.Display),
		Source:     &Source{},
		Filter:     []*AndConditions{},
		Mappings:   []*Mapping{},
		Dimensions: []string{},
		OrderBy:    []*OrderBy{},
		Aggregates: make(map[string][]*Mapping),
		Tags:       make(map[string][]*TagMapping),
		Union:      []*Union{},
		Join:       []*Join{},
		Sql:        utils.Trim(r.Sql),
	}
	cube.Sha1Name = Sha1Name(cube.Name)

	vals := strings.Split(r.Dimensions, SEP_LIST)
	for _, v := range vals {
		v = utils.Trim(v)
		if v != "" {
			cube.Dimensions = append(cube.Dimensions, v)
		}
	}

	for _, str := range r.OrderBy {
		orderBy := &OrderBy{
			Order: "ASC",
		}
		vals = strings.Split(str, SEP_LIST)
		if len(vals) == 0 {
			continue
		}
		orderBy.Field = utils.Trim(vals[0])
		if len(vals) >= 2 {
			orderBy.Order = utils.UpperTrim(vals[1])
		}
		cube.OrderBy = append(cube.OrderBy, orderBy)
	}

	if len(cube.OrderBy) == 0 && len(cube.Dimensions) > 0 {
		for _, v := range cube.Dimensions {
			cube.OrderBy = append(cube.OrderBy, &OrderBy{
				Field: v,
				Order: "ASC",
			})
		}
	}
	vals = strings.Split(r.Limit, SEP_LIST)
	if len(vals) > 1 {
		cube.Limit = &Limit{
			Limit:  1000,
			Offset: 0,
		}

		i, err := strconv.ParseInt(vals[0], 10, 64)
		if err == nil {
			cube.Limit.Limit = int(i)
		}

		if len(vals) >= 2 {
			i, err := strconv.ParseInt(vals[1], 10, 64)
			if err == nil {
				cube.Limit.Offset = int(i)
			}
		}
	}

	vals = strings.Split(r.Source, SEP_LIST)
	if len(vals) >= 1 {
		cube.Source.Type = utils.LowerTrim(vals[0])
	}
	if len(vals) >= 2 {
		cube.Source.Name = utils.Trim(vals[1])
	}

	//logger.Infof("r.Store = %s", r.Store)
	if r.Store != "" {
		vals = strings.Split(r.Store, SEP_LIST)
		if len(vals) > 0 {
			name := utils.Trim(vals[0])
			if name != "" {
				var alias string
				if len(vals) >= 2 {
					alias = utils.Trim(vals[1])
				}
				if alias == "" {
					alias = "t0"
				}
				cube.Store = &Store{
					Name:  name,
					Alias: alias,
				}
				if cube.Source.Type == SOURCE_MYSQL || cube.Source.Type == SOURCE_SQLITE {
					cube.Store.Sha1Name = cube.Store.Name
				} else {
					cube.Store.Sha1Name = Sha1Name(cube.Store.Name)
				}
			}
		}
	}
	//logger.Infof("cube.Store = %s", utils.Json(cube.Store))

	for _, orList := range r.Filter {
		andConds := &AndConditions{
			Conds: []*Condition{},
		}
		for _, andStr := range orList {
			and := strings.Split(andStr, SEP_EXP)
			if len(and) < 2 {
				continue
			}
			cond := &Condition{
				Field: utils.Trim(and[0]),
				Op:    utils.UpperTrim(and[1]),
			}
			if len(and) >= 3 {
				cond.Val = utils.Trim(and[2])
			}
			if len(and) >= 4 {
				cond.Val2 = utils.Trim(and[3])
			}
			andConds.Conds = append(andConds.Conds, cond)
		}
		cube.Filter = append(cube.Filter, andConds)
	}

	for _, v := range r.Mappings {
		vals := strings.Split(v, SEP_EXP)
		if len(vals) < 2 {
			continue
		}
		mapping := &Mapping{
			Alias:      utils.Trim(vals[0]),
			Expression: utils.Trim(vals[1]),
		}
		cube.Mappings = append(cube.Mappings, mapping)
	}

	for _, aggregate := range r.Aggregates {
		if len(aggregate) < 2 {
			continue
		}
		function := aggregate[0]
		mappings := []*Mapping{}
		for k, v := range aggregate {
			if k == 0 {
				continue
			}
			vals := strings.Split(v, SEP_EXP)
			if len(vals) == 0 {
				continue
			}
			mapping := &Mapping{
				Expression: utils.Trim(vals[0]),
			}
			if len(vals) >= 2 {
				mapping.Alias = utils.Trim(vals[1])
			}
			if mapping.Alias == "" {
				mapping.Alias = fmt.Sprintf("%s_%s", utils.LowerTrim(function), mapping.Expression)
			}
			mappings = append(mappings, mapping)
		}
		cube.Aggregates[function] = mappings
	}

	for tagName, mappings := range r.Tags {
		tagMappings := []*TagMapping{}
		for _, v := range mappings {
			vals := strings.Split(v, SEP_EXP)
			if len(vals) < 4 {
				continue
			}

			tagMapping := &TagMapping{
				TagVal: utils.Trim(vals[0]),
				Field:  utils.Trim(vals[1]),
				Op:     utils.UpperTrim(vals[2]),
				Val:    utils.Trim(vals[3]),
			}
			if len(vals) >= 5 {
				tagMapping.Val2 = utils.Trim(vals[4])
			}

			// TO DO: support more functions for tags
			if tagMapping.Op != "REGEXP" {
				return nil, errors.New("For now, only `regexp` supported for tags.")
			}
			reg, err := regexp.Compile(tagMapping.Val)
			if err != nil {
				logger.Error(err)
				return nil, err
			}
			tagMapping.IncludeRegexp = reg

			if tagMapping.Val2 != "" {
				reg, err = regexp.Compile(tagMapping.Val2)
				if err != nil {
					logger.Error(err)
					return nil, err
				}
				tagMapping.ExcludeRegexp = reg
			}
			tagMappings = append(tagMappings, tagMapping)
		}
		if len(tagMappings) > 0 {
			cube.Tags[tagName] = tagMappings
		}
	}

	for _, v := range r.Union {
		vals := strings.Split(v, SEP_LIST)
		if len(vals) >= 1 {
			union := &Union{
				Name:      utils.Trim(vals[0]),
				UnionType: UNION,
			}
			if len(vals) >= 2 {
				union.UnionType = utils.UpperTrim(vals[1])
			}
			if cube.Source.Type == SOURCE_MYSQL || cube.Source.Type == SOURCE_SQLITE {
				union.Sha1Name = union.Name
			} else {
				union.Sha1Name = Sha1Name(union.Name)
			}

			cube.Union = append(cube.Union, union)
		}
	}

	for _, v := range r.Join {
		join := v.Convert(cube.Source.Type)
		if join != nil {
			cube.Join = append(cube.Join, join)
		}
	}

	//logger.Infof("cube = %s", utils.Json(cube))
	if err := CheckCube(cube); err != nil {
		logger.Error(err)
		return nil, err
	}

	return cube, nil
}

func (j *JoinJson) Convert(sourceType string) *Join {
	vals := strings.Split(j.Store, SEP_LIST)
	if len(vals) == 0 {
		return nil
	}
	store := &Store{
		Name: utils.Trim(vals[0]),
	}
	if len(vals) >= 2 {
		store.Alias = utils.Trim(vals[1])
	}
	if store.Alias == "" {
		store.Alias = store.Name
	}
	if sourceType == SOURCE_MYSQL || sourceType == SOURCE_SQLITE {
		store.Sha1Name = store.Name
	} else {
		store.Sha1Name = Sha1Name(store.Name)
	}
	conds := []*Condition{}
	for _, v := range j.On {
		vals2 := strings.Split(v, SEP_EXP)
		if len(vals2) < 3 {
			continue
		}

		cond := &Condition{
			Field: utils.Trim(vals2[0]),
			Op:    utils.UpperTrim(vals2[1]),
			Val:   utils.Trim(vals2[2]),
		}
		if len(vals2) >= 4 {
			cond.Val2 = utils.Trim(vals2[3])
		}

		conds = append(conds, cond)
	}
	if len(conds) == 0 {
		return nil
	}

	join := &Join{
		Type:  utils.UpperTrim(j.Type),
		Store: store,
		On:    conds,
	}
	return join
}
