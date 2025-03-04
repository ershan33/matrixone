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

package disttae

import (
	"context"
	"time"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/pb/api"
	"github.com/matrixorigin/matrixone/pkg/pb/metadata"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/logtail"
)

func updatePartition(idx, primaryIdx int, tbl *table, ts timestamp.Timestamp,
	ctx context.Context, op client.TxnOperator, db *DB,
	mvcc MVCC, dn DNStore, req api.SyncLogTailReq) error {
	reqs, err := genLogTailReq(dn, req)
	if err != nil {
		return err
	}
	logTails, err := getLogTail(ctx, op, reqs)
	if err != nil {
		return err
	}
	for i := range logTails {
		if err := consumeLogTail(idx, primaryIdx, tbl, ts, ctx, db, mvcc, logTails[i]); err != nil {
			logutil.Errorf("consume %d-%s logtail error: %v\n", tbl.tableId, tbl.tableName, err)
			return err
		}
	}
	return nil
}

func getLogTail(ctx context.Context, op client.TxnOperator, reqs []txn.TxnRequest) ([]*api.SyncLogTailResp, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	result, err := op.Read(ctx, reqs)
	if err != nil {
		return nil, err
	}
	logTails := make([]*api.SyncLogTailResp, len(result.Responses))
	for i, resp := range result.Responses {
		logTails[i] = new(api.SyncLogTailResp)
		if err := types.Decode(resp.CNOpResponse.Payload, logTails[i]); err != nil {
			return nil, err
		}
	}
	return logTails, nil
}

func consumeLogTail(idx, primaryIdx int, tbl *table, ts timestamp.Timestamp,
	ctx context.Context, db *DB, mvcc MVCC, logTail *api.SyncLogTailResp) (err error) {
	var entries []*api.Entry

	if entries, err = logtail.LoadCheckpointEntries(
		ctx,
		logTail.CkpLocation,
		tbl.tableId,
		tbl.tableName,
		tbl.db.databaseId,
		tbl.db.databaseName,
		tbl.db.fs); err != nil {
		return
	}
	for _, e := range entries {
		if err = consumeEntry(idx, primaryIdx, tbl, ts, ctx,
			db, mvcc, e); err != nil {
			return
		}
	}

	for i := 0; i < len(logTail.Commands); i++ {
		if err = consumeEntry(idx, primaryIdx, tbl, ts, ctx,
			db, mvcc, logTail.Commands[i]); err != nil {
			return
		}
	}
	return nil
}

func consumeEntry(idx, primaryIdx int, tbl *table, ts timestamp.Timestamp,
	ctx context.Context, db *DB, mvcc MVCC, e *api.Entry) error {
	if e.EntryType == api.Entry_Insert {
		if isMetaTable(e.TableName) {
			vec, err := vector.ProtoVectorToVector(e.Bat.Vecs[catalog.BLOCKMETA_ID_IDX+MO_PRIMARY_OFF])
			if err != nil {
				return err
			}
			timeVec, err := vector.ProtoVectorToVector(e.Bat.Vecs[catalog.BLOCKMETA_COMMITTS_IDX+MO_PRIMARY_OFF])
			if err != nil {
				return err
			}
			vs := vector.MustTCols[uint64](vec)
			timestamps := vector.MustTCols[types.TS](timeVec)
			for i, v := range vs {
				if err := tbl.parts[idx].DeleteByBlockID(ctx, timestamps[i].ToTimestamp(), v); err != nil {
					if !moerr.IsMoErrCode(err, moerr.ErrTxnWriteConflict) {
						return err
					}
				}
			}
			return db.getMetaPartitions(e.TableName)[idx].Insert(ctx, -1, e.Bat, false)
		}
		switch e.TableId {
		case catalog.MO_TABLES_ID:
			bat, _ := batch.ProtoBatchToBatch(e.Bat)
			tbl.db.txn.catalog.InsertTable(bat)
		case catalog.MO_DATABASE_ID:
			bat, _ := batch.ProtoBatchToBatch(e.Bat)
			tbl.db.txn.catalog.InsertDatabase(bat)
		case catalog.MO_COLUMNS_ID:
			bat, _ := batch.ProtoBatchToBatch(e.Bat)
			tbl.db.txn.catalog.InsertColumns(bat)
		}
		if primaryIdx >= 0 {
			return mvcc.Insert(ctx, MO_PRIMARY_OFF+primaryIdx, e.Bat, false)
		}
		return mvcc.Insert(ctx, primaryIdx, e.Bat, false)
	}
	if isMetaTable(e.TableName) {
		return db.getMetaPartitions(e.TableName)[idx].Delete(ctx, e.Bat)
	}
	switch e.TableId {
	case catalog.MO_TABLES_ID:
		bat, _ := batch.ProtoBatchToBatch(e.Bat)
		tbl.db.txn.catalog.DeleteTable(bat)
	case catalog.MO_DATABASE_ID:
		bat, _ := batch.ProtoBatchToBatch(e.Bat)
		tbl.db.txn.catalog.DeleteDatabase(bat)
	}
	return mvcc.Delete(ctx, e.Bat)
}

func genSyncLogTailReq(have, want timestamp.Timestamp, databaseId,
	tableId uint64) api.SyncLogTailReq {
	return api.SyncLogTailReq{
		CnHave: &have,
		CnWant: &want,
		Table: &api.TableID{
			DbId: databaseId,
			TbId: tableId,
		},
	}
}

func genLogTailReq(dn DNStore, req api.SyncLogTailReq) ([]txn.TxnRequest, error) {
	payload, err := types.Encode(req)
	if err != nil {
		return nil, err
	}
	reqs := make([]txn.TxnRequest, len(dn.Shards))
	for i, info := range dn.Shards {
		reqs[i] = txn.TxnRequest{
			CNRequest: &txn.CNOpRequest{
				OpCode:  uint32(api.OpCode_OpGetLogTail),
				Payload: payload,
				Target: metadata.DNShard{
					DNShardRecord: metadata.DNShardRecord{
						ShardID: info.ShardID,
					},
					ReplicaID: info.ReplicaID,
					Address:   dn.ServiceAddress,
				},
			},
			Options: &txn.TxnRequestOptions{
				RetryCodes: []int32{
					// dn shard not found
					int32(moerr.ErrDNShardNotFound),
				},
				RetryInterval: int64(time.Second),
			},
		}
	}
	return reqs, nil
}
