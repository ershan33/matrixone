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

package tree

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQueryType(t *testing.T) {
	type fields struct {
		types map[StatementType]string
	}
	type args struct {
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "normal",
			fields: fields{
				types: map[StatementType]string{
					&Select{}:             QueryTypeDQL,
					&SelectClause{}:       QueryTypeDQL,
					&MoDump{}:             QueryTypeDQL,
					&ShowProcessList{}:    QueryTypeDQL,
					&ShowTableStatus{}:    QueryTypeDQL,
					&ShowCreateDatabase{}: QueryTypeDQL,
					&ShowCreateView{}:     QueryTypeDQL,
					&ShowCreateTable{}:    QueryTypeDQL,
					&ShowAccounts{}:       QueryTypeDQL,
					&ShowCollation{}:      QueryTypeDQL,
					&ShowDatabases{}:      QueryTypeDQL,
					&ShowErrors{}:         QueryTypeDQL,
					&ShowFunctionStatus{}: QueryTypeDQL,
					&ShowGrants{}:         QueryTypeDQL,
					&ShowIndex{}:          QueryTypeDQL,
					&ShowLocks{}:          QueryTypeDQL,
					&ShowNodeList{}:       QueryTypeDQL,
					&ShowStatus{}:         QueryTypeDQL,
					&ShowTableNumber{}:    QueryTypeDQL,
					&ShowTables{}:         QueryTypeDQL,
					&ShowTarget{}:         QueryTypeDQL,
					&ShowVariables{}:      QueryTypeDQL,
					&ShowWarnings{}:       QueryTypeDQL,
					&ValuesStatement{}:    QueryTypeDQL,
					&With{}:               QueryTypeDQL,
					// DDL
					&CreateDatabase{}:      QueryTypeDDL,
					&CreateTable{}:         QueryTypeDDL,
					&CreateView{}:          QueryTypeDDL,
					&CreateIndex{}:         QueryTypeDDL,
					&CreateFunction{}:      QueryTypeDDL,
					&AlterView{}:           QueryTypeDDL,
					&AlterDataBaseConfig{}: QueryTypeDDL,
					&DropDatabase{}:        QueryTypeDDL,
					&DropTable{}:           QueryTypeDDL,
					&DropView{}:            QueryTypeDDL,
					&DropIndex{}:           QueryTypeDDL,
					&DropFunction{}:        QueryTypeDDL,
					&TruncateTable{}:       QueryTypeDDL,
					// DCL
					&CreateAccount{}:   QueryTypeDCL,
					&CreateRole{}:      QueryTypeDCL,
					&CreateUser{}:      QueryTypeDCL,
					&Grant{}:           QueryTypeDCL,
					&GrantPrivilege{}:  QueryTypeDCL,
					&GrantProxy{}:      QueryTypeDCL,
					&GrantRole{}:       QueryTypeDCL,
					&Revoke{}:          QueryTypeDCL,
					&RevokePrivilege{}: QueryTypeDCL,
					&RevokeRole{}:      QueryTypeDCL,
					&AlterAccount{}:    QueryTypeDCL,
					&AlterUser{}:       QueryTypeDCL,
					&DropAccount{}:     QueryTypeDCL,
					&DropRole{}:        QueryTypeDCL,
					&DropUser{}:        QueryTypeDCL,
					// TCL
					&BeginTransaction{}:    QueryTypeTCL,
					&CommitTransaction{}:   QueryTypeTCL,
					&RollbackTransaction{}: QueryTypeTCL,
					// Other
					&AnalyzeStmt{}:    QueryTypeOth,
					&ExplainAnalyze{}: QueryTypeOth,
					&ExplainFor{}:     QueryTypeOth,
					&ExplainStmt{}:    QueryTypeOth,
					&SetVar{}:         QueryTypeOth,
					&SetDefaultRole{}: QueryTypeOth,
					&SetRole{}:        QueryTypeOth,
					&SetPassword{}:    QueryTypeOth,
					&Declare{}:        QueryTypeOth,
					&Do{}:             QueryTypeOth,
					&TableFunction{}:  QueryTypeOth,
					&Use{}:            QueryTypeOth,
					&PrepareStmt{}:    QueryTypeOth,
					&Execute{}:        QueryTypeOth,
					&Deallocate{}:     QueryTypeOth,
					&Kill{}:           QueryTypeOth,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			for stmt, queryType := range tt.fields.types {
				require.Equalf(t1, queryType, stmt.GetQueryType(), "statement_type: %s", stmt.GetStatementType())
			}
		})
	}

}
