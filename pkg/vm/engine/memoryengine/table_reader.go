// Copyright 2022 Matrix Origin
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

package memoryengine

import (
	"context"
	"encoding/binary"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/pb/metadata"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/mheap"
)

type TableReader struct {
	ctx         context.Context
	engine      *Engine
	txnOperator client.TxnOperator
	iterInfos   []IterInfo
}

type IterInfo struct {
	Shard  Shard
	IterID ID
}

func (t *Table) NewReader(
	ctx context.Context,
	parallel int,
	expr *plan.Expr,
	shardIDs [][]byte,
) (
	readers []engine.Reader,
	err error,
) {

	readers = make([]engine.Reader, parallel)

	var shards []Shard
	if len(shardIDs) == 0 {
		// all
		var err error
		shards, err = t.engine.allShards()
		if err != nil {
			return nil, err
		}

	} else {
		// some
		idSet := make(map[uint64]bool)
		for _, bs := range shardIDs {
			id := binary.LittleEndian.Uint64(bs)
			idSet[id] = true
		}
		clusterDetails, err := t.engine.getClusterDetails()
		if err != nil {
			return nil, err
		}
		for _, store := range clusterDetails.DNStores {
			for _, shard := range store.Shards {
				if !idSet[shard.ShardID] {
					continue
				}
				shards = append(shards, Shard{
					DNShardRecord: metadata.DNShardRecord{
						ShardID: shard.ShardID,
					},
					ReplicaID: shard.ReplicaID,
					Address:   store.ServiceAddress,
				})
			}
		}
	}

	resps, err := DoTxnRequest[NewTableIterResp](
		ctx,
		t.engine,
		t.txnOperator.Read,
		theseShards(shards),
		OpNewTableIter,
		NewTableIterReq{
			TableID: t.id,
			Expr:    expr,
		},
	)
	if err != nil {
		return nil, err
	}

	iterInfoSets := make([][]IterInfo, parallel)
	for i, resp := range resps {
		if resp.IterID == emptyID {
			continue
		}
		iterInfo := IterInfo{
			Shard:  shards[i],
			IterID: resp.IterID,
		}
		iterInfoSets[i%parallel] = append(iterInfoSets[i%parallel], iterInfo)
	}

	for i, set := range iterInfoSets {
		if len(set) == 0 {
			readers[i] = new(TableReader)
			continue
		}
		reader := &TableReader{
			engine:      t.engine,
			txnOperator: t.txnOperator,
			ctx:         ctx,
			iterInfos:   set,
		}
		readers[i] = reader
	}

	return
}

var _ engine.Reader = new(TableReader)

func (t *TableReader) Read(colNames []string, plan *plan.Expr, mh *mheap.Mheap) (*batch.Batch, error) {
	if t == nil {
		return nil, nil
	}

	for {

		if len(t.iterInfos) == 0 {
			return nil, nil
		}

		resps, err := DoTxnRequest[ReadResp](
			t.ctx,
			t.engine,
			t.txnOperator.Read,
			thisShard(t.iterInfos[0].Shard),
			OpRead,
			ReadReq{
				IterID:   t.iterInfos[0].IterID,
				ColNames: colNames,
			},
		)
		if err != nil {
			return nil, err
		}

		resp := resps[0]

		if resp.Batch == nil {
			// no more
			t.iterInfos = t.iterInfos[1:]
			continue
		}

		return resp.Batch, nil
	}

}

func (t *TableReader) Close() error {
	if t == nil {
		return nil
	}
	for _, info := range t.iterInfos {
		_, err := DoTxnRequest[CloseTableIterResp](
			t.ctx,
			t.engine,
			t.txnOperator.Read,
			thisShard(info.Shard),
			OpCloseTableIter,
			CloseTableIterReq{
				IterID: info.IterID,
			},
		)
		_ = err // ignore error
	}
	return nil
}

func (t *Table) Ranges(ctx context.Context) ([][]byte, error) {
	// return encoded shard ids
	clusterDetails, err := t.engine.getClusterDetails()
	if err != nil {
		return nil, err
	}
	nodes := clusterDetails.DNStores
	shards := make([][]byte, 0, len(nodes))
	for _, node := range nodes {
		for _, shard := range node.Shards {
			id := make([]byte, 8)
			binary.LittleEndian.PutUint64(id, shard.ShardID)
			shards = append(shards, id)
		}
	}
	return shards, nil
}