// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package db

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/xormplus/xorm"

	"github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/apistructs"
)

// PipelineCmsNs 配置管理命名空间
type PipelineCmsNs struct {
	ID uint64 `json:"id" xorm:"pk autoincr"`

	PipelineSource apistructs.PipelineSource `json:"pipelineSource"`

	Ns string `json:"ns"`

	TimeCreated *time.Time `json:"timeCreated,omitempty" xorm:"created"`
	TimeUpdated *time.Time `json:"timeUpdated,omitempty" xorm:"updated"`
}

func (PipelineCmsNs) TableName() string {
	return "dice_pipeline_cms_ns"
}

// PipelineCmsConfig 配置管理命名空间下的具体配置
type PipelineCmsConfig struct {
	ID uint64 `json:"id" xorm:"pk autoincr"`

	NsID uint64 `json:"nsID"`

	Key   string `json:"key"`
	Value string `json:"value"`

	Encrypt *bool `json:"encrypt"`

	Type string `json:"type"`

	Extra PipelineCmsConfigExtra `json:"extra" xorm:"json"`

	TimeCreated *time.Time `json:"timeCreated,omitempty" xorm:"created"`
	TimeUpdated *time.Time `json:"timeUpdated,omitempty" xorm:"updated"`
}

// BeforeSet is invoked before FromDB
// order: get value from db -> invoke BeforeSet -> invoke FromDB -> struct
func (c PipelineCmsConfig) BeforeSet(fieldName string, cell xorm.Cell) {
	switch fieldName {
	case "type":
		// NULL -> kv
		if reflect.Indirect(reflect.ValueOf(cell)).IsNil() {
			*cell = ConfigTypeKV
		}
	case "extra":
		// NULL -> ""
		// set to "" to enable PipelineCmsConfigExtra.FromDB
		if reflect.Indirect(reflect.ValueOf(cell)).IsNil() {
			*cell = ""
		}
	}
}

type PipelineCmsConfigExtra struct {
	// Operations 从数据库取出时保证不为 nil
	Operations *pb.PipelineCmsConfigOperations `json:"operations"`
	// Comment 注释
	Comment string `json:"comment"`
	// From 配置项来源，可为空。例如：证书管理同步
	From string `json:"from"`
}

// FromDB 处理 operations 默认值，老数据无需通过 dbmigration 赋值
func (extra *PipelineCmsConfigExtra) FromDB(b []byte) error {
	if len(b) > 0 {
		if err := json.Unmarshal(b, extra); err != nil {
			return err
		}
	}
	if extra.Operations == nil {
		extra.Operations = &DefaultOperationsForKV
	}
	return nil
}

// ToDB 为 operations 赋默认值
func (extra *PipelineCmsConfigExtra) ToDB() ([]byte, error) {
	if extra.Operations == nil {
		extra.Operations = &DefaultOperationsForKV
	}
	return json.Marshal(extra)
}

func (PipelineCmsConfig) TableName() string {
	return "dice_pipeline_cms_configs"
}

func (c PipelineCmsConfig) Equal(another PipelineCmsConfig) bool {
	return c.NsID == another.NsID &&
		c.Key == another.Key &&
		c.Value == another.Value &&
		reflect.DeepEqual(c.Encrypt, another.Encrypt) &&
		c.Type == another.Type &&
		reflect.DeepEqual(c.Extra, another.Extra)
}
