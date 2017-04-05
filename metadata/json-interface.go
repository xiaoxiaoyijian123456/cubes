package metadata

const (
	SEP_LIST = ","
	SEP_EXP  = ";"
)

type ReportJson struct {
	Report     []string        `json:"report,omitempty"`
	Comment    string          `json:"_comment,omitempty"`
	Cube       *CubeJson       `json:"cube,omitempty"`
	Cubes      *CubesJson      `json:"cubes,omitempty"`
	CubesGroup *CubesGroupJson `json:"cubes_group,omitempty"`
}
type CubesJson struct {
	Comment  string      `json:"_comment,omitempty"`
	CubeList []*CubeJson `json:"cube_list,omitempty"`
}

type CubesGroupJson struct {
	Comment   string       `json:"_comment,omitempty"`
	CubesList []*CubesJson `json:"cubes_list,omitempty"`
}

type CubeJson struct {
	Name       string              `json:"name,omitempty"`
	Comment    string              `json:"_comment,omitempty"`
	Display    string              `json:"display,omitempty"` // 在结果中原样返回，用户前段显示数据用
	Source     string              `json:"source,omitempty"`
	Store      string              `json:"store,omitempty"`
	Join       []*JoinJson         `json:"join,omitempty"`
	Filter     [][]string          `json:"filter,omitempty"`
	Mappings   []string            `json:"mappings,omitempty"`
	Dimensions string              `json:"dimensions,omitempty"`
	OrderBy    []string            `json:"orderby,omitempty"`
	Limit      string              `json:"limit,omitempty"`
	Aggregates [][]string          `json:"aggregates,omitempty"`
	Tags       map[string][]string `json:"tags,omitempty"`
	Union      []string            `json:"Union,omitempty"`
	Sql        string              `json:"sql,omitempty"`
}

type JoinJson struct {
	Type  string   `json:"type,omitempty"`
	Store string   `json:"store,omitempty"`
	On    []string `json:"on,omitempty"`
}
