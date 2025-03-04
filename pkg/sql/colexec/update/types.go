// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package update

import (
	"github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

type Argument struct {
	Ts           uint64
	AffectedRows uint64
	UpdateCtxs   []*UpdateCtx
	TableDefVec  []*plan.TableDef
	Engine       engine.Engine
	DB           []engine.Database
	TableID      []uint64
	DBName       []string
	TblName      []string
	HasAutoCol   []bool
}

func (arg *Argument) Free(proc *process.Process, pipelineFailed bool) {
}

type UpdateCtx struct {
	HideKey     string
	HideKeyIdx  int32
	UpdateAttrs []string
	OtherAttrs  []string
	OrderAttrs  []string
	TableSource engine.Relation
	// for not index table
	CPkeyColDef        *plan.ColDef
	IsIndexTableUpdate bool
	// for index table
	UniqueIndexPos    []int
	SecondaryIndexPos []int
	IndexParts        []string
	ClusterByDef      *plan.ClusterByDef
}
