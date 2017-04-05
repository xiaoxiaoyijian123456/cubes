package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/xiaoxiaoyijian123456/cubes/utils"
	"sync"
)

func NewReport() *Report {
	return &Report{
		Report: []string{},
		Cubes:  utils.NewMapData(),
	}
}

const MAX_LOOP = 200

func (r *Report) Execute() (map[string]*CubeReport, error) {
	cubes_result := utils.NewMapData()

	cube_tasks := make(map[string]*Cube)
	for k, v := range r.Cubes.Copy() {
		cube, ok := v.(*Cube)
		if !ok {
			err := errors.New("Map data should return cube.")
			logger.Error(err)
			return nil, err
		}
		cubeName, ok := k.(string)
		if !ok {
			err := errors.New("Map key should return string.")
			logger.Error(err)
			return nil, err
		}
		cube_tasks[cubeName] = cube
	}

	loop := 0

	for {
		if len(cube_tasks) == 0 {
			break
		}

		logger.Infof("LOOP: %d", loop)
		loop++
		if loop >= MAX_LOOP {
			err := errors.New(fmt.Sprintf("Has invalid cubes: %v", utils.Json(cube_tasks)))
			logger.Error(err)
			return nil, err
		}

		batch_tasks := []*Cube{}
		for _, c := range cube_tasks {
			logger.Infof("======================processing cube: %s", c.Name)

			if v := cubes_result.Get(c.Name); v != nil {
				delete(cube_tasks, c.Name)
				continue
			}

			needWait := false
			if c.Source.Type == SOURCE_CUBE {
				if c.Store != nil {
					if v := cubes_result.Get(c.Store.Name); v == nil {
						needWait = true
					}
				}
				for _, union := range c.Union {
					if v := cubes_result.Get(union.Name); v == nil {
						needWait = true
					}
				}
			}
			if needWait {
				logger.Infof("======================cube: %s need wait, cube.store=%s, cube.union=%s.", c.Name, utils.Json(c.Store), utils.Json(c.Union))
				continue
			}

			batch_tasks = append(batch_tasks, c)
		}

		var wg sync.WaitGroup
		errorsMap := utils.NewMapData()
		for _, c := range batch_tasks {
			wg.Add(1)
			go func(cube *Cube) {
				defer wg.Done()

				if err := cube.Execute(); err != nil {
					errorsMap.Set(cube.Name, err)
					logger.Error(err)
					return
				}
				report, err := cube.GetReport()
				if err != nil {
					errorsMap.Set(cube.Name, err)
					logger.Error(err)
					return
				}

				cubes_result.Set(cube.Name, report)
			}(c)
		}
		wg.Wait()
		batch_tasks = []*Cube{}

		if errorsMap.Len() > 0 {
			var buffer bytes.Buffer
			for k, v := range errorsMap.Copy() {
				buffer.WriteString(fmt.Sprintf("%v: %v", k, v))
			}
			err := errors.New(fmt.Sprintf("Errors found in cube execution, errors:%v", buffer.String()))
			logger.Error(err)
			return nil, err
		}
	}

	ret := make(map[string]*CubeReport)
	for _, v := range r.Report {
		if report := cubes_result.Get(v); report != nil {
			cube_result, ok := report.(*CubeReport)
			if !ok {
				err := errors.New(fmt.Sprintf("Report cube[%s] not found.", v))
				logger.Error(err)
				return nil, err
			}
			ret[v] = cube_result
		}
	}
	return ret, nil
}
