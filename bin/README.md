# 报表分析配置文件格式
# 概览
* JSON格式
* 默认文件名：report.json
* 示例：
```json
{
  "report":["select_cube"],
  "_comment": "example report for mysql select.",
  "cube" : {
      "name": "select_cube",
      "_comment": "example cube for mysql select",
      "source": "mysql",
      "store": "skyline.zhizuan_campaign_rpt_daily",
      "filter": [
        ["client_id;=;1",
          "record_on;between;2017-03-06;2017-03-12"
        ]
      ],
      "orderby":["record_on, DESC"],
      "limit":"1,0"
  }
}
```
# 格式描述
## reports
* 格式定义：
```json
{
  "report": ["<cube_name_1>", "<cube_name_2>",...],
  "_comment": "<description for report>",
  "cube" : <cube structure>,
  "cubes" : <cubes structure>,
  "cubes_group" : <cubes_group structure>
}
```
* 同一个配置文件中可以输出一个或者多个透视(cube)结果
* 同一个配置文件中，cube, cubes, cubes_group至少需要定义一项，根据实际选择：
  1. cube用于定义单个透视
  2. cubes用于定义单组透视
  3. cubes_group用于将透视分成多组
* 同一个配置文件中，可以定义多层透视（即Cube-B的数据来源是Cube-A)
* 返回结果格式定义：
```json
{
  "<cube_name_1>":[
    {
      <field1>:<val1>,
      <field2>:<val2>,
      ..
    },
    {
      <field1>:<val1_2>,
      <field2>:<val2_2>,
      ..
    },    
    ...
  ],
  "<cube_name_2>":[
    {
      <field_1>:<val1>,
      <field_2>:<val2>,
      ..
    },
    {
      <field_1>:<val1_2>,
      <field_2>:<val2_2>,
      ..
    },    
    ...
  ],  
}
```
## 透视单元（cube）
### 格式定义：
* 数据来源：CSV / JSON
```json
{
      "name": "<cube_name>",
      "_comment": "<description for cube>",
      "source": "csv|json,<csv/json_file_path>",
      "dimensions":"<field1>,<field2>,...<tag_name1>,<tag_name2>,...",
      "mapping":[
        "<field_alias1>;<expression_1>",
        "<field_alias2>;<expression_2>",
        ...
      ], 
      "filter": [
        ["<field_name>;<op>;<value>",
          "record_on;between;2017-03-06;2017-03-12"
        ]
      ],           
      "orderby":[
        "<field1>,DESC|ASC",
        "<field2>,DESC|ASC",
        ...
      ],
      "limit":"<limit>,<offset>",
      "aggregates": [
        ["<func>","<field1>;<field1_alias>","<field2>;<field2_alias>,..."]
      ],
      "tags": {
        "<tag_name1>": [
          "<tag_name1_val1>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          "<tag_name1_val2>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          ...
        ],
        "<tag_name2>": [
          "<tag_name2_val1>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          "<tag_name2_val2>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          ...
        ],
      }
},
```
* 数据来源：mysql
```json
{
      "name": "<cube_name>",
      "_comment": "<description for cube>",
      "source": "mysql",
      "store": "<table_name>,<table_alias>",
      "dimensions":"<field1>,<field2>,...<tag_name1>,<tag_name2>,...",
      "mapping":[
        "<field_alias1>;<expression_1>",
        "<field_alias2>;<expression_2>",
        ...
      ],      
      "orderby":[
        "<field1>,DESC|ASC",
        "<field2>,DESC|ASC",
        ...
      ],
      "limit":"<limit>,<offset>",
      "aggregates": [
        ["<func>","<field1>;<field1_alias>","<field2>;<field2_alias>,..."]
      ],
      "tags": {
        "<tag_name1>": [
          "<tag_name1_val1>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          "<tag_name1_val2>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          ...
        ],
        "<tag_name2>": [
          "<tag_name2_val1>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          "<tag_name2_val2>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          ...
        ],
      }
},
```
* 数据来源：sqlite
```json
{
      "name": "<cube_name>",
      "_comment": "<description for cube>",
      "source": "sqlite,<sqlite_db_path>",
      "store": "<table_name>,<table_alias>",
      "dimensions":"<field1>,<field2>,...<tag_name1>,<tag_name2>,...",
      "mapping":[
        "<field_alias1>;<expression_1>",
        "<field_alias2>;<expression_2>",
        ...
      ],      
      "orderby":[
        "<field1>,DESC|ASC",
        "<field2>,DESC|ASC",
        ...
      ],
      "limit":"<limit>,<offset>",
      "aggregates": [
        ["<func>","<field1>;<field1_alias>","<field2>;<field2_alias>,..."]
      ],
      "tags": {
        "<tag_name1>": [
          "<tag_name1_val1>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          "<tag_name1_val2>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          ...
        ],
        "<tag_name2>": [
          "<tag_name2_val1>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          "<tag_name2_val2>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          ...
        ],
      }
},
```
* 数据来源：cube (多层透视)
```json
{
      "name": "<cube_name>",
      "_comment": "<description for cube>",
      "source": "cube",
      "store": "<cube_name>,<cube_alias>",
      "dimensions":"<field1>,<field2>,...<tag_name1>,<tag_name2>,...",
      "mapping":[
        "<field_alias1>;<expression_1>",
        "<field_alias2>;<expression_2>",
        ...
      ],      
      "orderby":[
        "<field1>,DESC|ASC",
        "<field2>,DESC|ASC",
        ...
      ],
      "limit":"<limit>,<offset>",
      "aggregates": [
        ["<func>","<field1_expression>;<field1_alias>","<field2_expression>;<field2_alias>,..."]
      ],
      "tags": {
        "<tag_name1>": [
          "<tag_name1_val1>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          "<tag_name1_val2>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          ...
        ],
        "<tag_name2>": [
          "<tag_name2_val1>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          "<tag_name2_val2>;<fieldX>;REGEXP;<include_regexp>;<exclude_regexp>",
          ...
        ],
      }
},
```
### cube name
同一个配置文件中，cube name不能重复
### cube comment
cube的备注信息
### source（数据来源）
支持csv/json/mysql/sqlite/cube(多层透视)
### store (对应数据库中的table) , store_alias
store_alias可选
### mapping（字段映射）
可以根据表达式定义新的字段，注意字段别名和对应的表达式之间是以分号（;)分隔。
注意：只有数据来源为MYSQL时，使用MYSQL的表达式，其他情况下默认使用SQLITE的表达式。
### dimensions（维度字段）
定义需要输出的字段列表
### order by
定义结果输出的排序，支持多个结果字段的ASC/DESC排序
### aggregates
定义group by操作，支持的函数有SUM, COUNT, MIN, MAX, AVG等, 针对的字段可以是字段表达式，操作的结果字段可以起个别名.
### tags（标签）
针对特定字段，使用正则表达式匹配出新的标签字段
### limit（限制返回记录数）
定义limit操作，offset可选，默认0.











