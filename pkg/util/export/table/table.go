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

package table

import (
	"context"
	"fmt"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"strings"
	"sync"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/util/batchpipe"
)

const ExternalFilePath = "__mo_filepath"

type CsvOptions struct {
	FieldTerminator rune // like: ','
	EncloseRune     rune // like: '"'
	Terminator      rune // like: '\n'
}

var CommonCsvOptions = &CsvOptions{
	FieldTerminator: ',',
	EncloseRune:     '"',
	Terminator:      '\n',
}

type ColType int

const (
	TDatetime ColType = iota
	TUint64
	TInt64
	TFloat64
	TJson
	TText
	TVarchar
)

func (c *ColType) ToType() types.Type {
	switch *c {
	case TDatetime:
		typ := types.T_datetime.ToType()
		typ.Precision = 6
		return typ
	case TUint64:
		return types.T_uint64.ToType()
	case TInt64:
		return types.T_int64.ToType()
	case TFloat64:
		return types.T_float64.ToType()
	case TJson:
		return types.T_json.ToType()
	case TText:
		return types.T_text.ToType()
	case TVarchar:
		return types.T_varchar.ToType()
	default:
		panic("not support ColType")
	}
}

type Column struct {
	Name    string
	Type    string
	Default string
	Comment string
	Alias   string // only use in view
	ColType ColType
}

// ToCreateSql return column scheme in create sql
//   - case 1: `column_name` varchar(36) DEFAULT "def_val" COMMENT "what am I, with default."
//   - case 2: `column_name` varchar(36) NOT NULL COMMENT "what am I. Without default, SO NOT NULL."
func (col *Column) ToCreateSql(ctx context.Context) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("`%s` %s ", col.Name, col.Type))
	if strings.ToUpper(col.Type) == "JSON" {
		sb.WriteString("NOT NULL ")
		if len(col.Default) == 0 {
			panic(moerr.NewNotSupported(ctx, "json column need default in csv, but not in schema"))
		}
	} else if len(col.Default) > 0 {
		sb.WriteString(fmt.Sprintf("DEFAULT %q ", col.Default))
	} else {
		sb.WriteString("NOT NULL ")
	}
	sb.WriteString(fmt.Sprintf("COMMENT %q", col.Comment))
	return sb.String()
}

var _ batchpipe.HasName = (*Table)(nil)

var NormalTableEngine = "TABLE"
var ExternalTableEngine = "EXTERNAL"

type Table struct {
	Account          string
	Database         string
	Table            string
	Columns          []Column
	PrimaryKeyColumn []Column
	Engine           string
	Comment          string
	// PathBuilder help to desc param 'infile'
	PathBuilder PathBuilder
	// AccountColumn help to split data in account's filepath
	AccountColumn *Column
	// TableOptions default is nil, see GetTableOptions
	TableOptions TableOptions
	// SupportUserAccess default false. if true, user account can access.
	SupportUserAccess bool
	// SupportConstAccess default false. if true, use Table.Account
	SupportConstAccess bool
}

func (tbl *Table) Clone() *Table {
	t := &Table{}
	*t = *tbl
	return t
}

func (tbl *Table) GetName() string {
	return tbl.Table
}
func (tbl *Table) GetDatabase() string {
	return tbl.Database
}

// GetIdentify return identify like database.table
func (tbl *Table) GetIdentify() string {
	return fmt.Sprintf("%s.%s", tbl.Database, tbl.Table)
}

type TableOptions interface {
	FormatDdl(ddl string) string
	// GetCreateOptions return option for `create {option}table`, which should end with ' '
	GetCreateOptions() string
	GetTableOptions(PathBuilder) string
}

func (tbl *Table) ToCreateSql(ctx context.Context, ifNotExists bool) string {

	TableOptions := tbl.GetTableOptions(ctx)

	const newLineCharacter = ",\n"
	sb := strings.Builder{}
	// create table
	sb.WriteString("CREATE ")
	switch strings.ToUpper(tbl.Engine) {
	case ExternalTableEngine:
		sb.WriteString(TableOptions.GetCreateOptions())
	default:
		panic(moerr.NewInternalError(ctx, "NOT support engine: %s", tbl.Engine))
	}
	sb.WriteString("TABLE ")
	if ifNotExists {
		sb.WriteString("IF NOT EXISTS ")
	}
	// table name
	sb.WriteString(fmt.Sprintf("`%s`.`%s`(", tbl.Database, tbl.Table))
	// columns
	for idx, col := range tbl.Columns {
		if idx > 0 {
			sb.WriteString(newLineCharacter)
		} else {
			sb.WriteRune('\n')
		}
		sb.WriteString(col.ToCreateSql(ctx))
	}
	// primary key
	if len(tbl.PrimaryKeyColumn) > 0 && tbl.Engine != ExternalTableEngine {
		sb.WriteString(newLineCharacter)
		sb.WriteString("PRIMARY KEY (")
		for idx, col := range tbl.PrimaryKeyColumn {
			if idx > 0 {
				sb.WriteString(`, `)
			}
			sb.WriteString(fmt.Sprintf("`%s`", col.Name))
		}
		sb.WriteString(`)`)
	}
	sb.WriteString("\n)")
	sb.WriteString(TableOptions.GetTableOptions(tbl.PathBuilder))

	return sb.String()
}

func (tbl *Table) GetTableOptions(ctx context.Context) TableOptions {
	if tbl.TableOptions != nil {
		return tbl.TableOptions
	}
	return GetOptionFactory(ctx, tbl.Engine)(tbl.Database, tbl.Table, tbl.Account)
}

type ViewOption func(view *View)

func (opt ViewOption) Apply(view *View) {
	opt(view)
}

type WhereCondition interface {
	String() string
}

type View struct {
	Database    string
	Table       string
	OriginTable *Table
	Columns     []Column
	Condition   WhereCondition
	// SupportUserAccess default false. if true, user account can access.
	SupportUserAccess bool
}

func WithColumn(c Column) ViewOption {
	return ViewOption(func(v *View) {
		v.Columns = append(v.Columns, c)
	})
}

func SupportUserAccess(support bool) ViewOption {
	return ViewOption(func(v *View) {
		v.SupportUserAccess = support
	})
}

func (tbl *View) ToCreateSql(ctx context.Context, ifNotExists bool) string {
	sb := strings.Builder{}
	// create table
	sb.WriteString("CREATE VIEW ")
	if ifNotExists {
		sb.WriteString("IF NOT EXISTS ")
	}
	// table name
	sb.WriteString(fmt.Sprintf("`%s`.`%s` as ", tbl.Database, tbl.Table))
	sb.WriteString("select ")
	// columns
	for idx, col := range tbl.Columns {
		if idx > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("`%s`", col.Name))
		if len(col.Alias) > 0 {
			sb.WriteString(fmt.Sprintf(" as `%s`", col.Alias))
		}
	}
	if tbl.OriginTable.Engine == ExternalTableEngine {
		sb.WriteString(fmt.Sprintf(", `%s`", ExternalFilePath))
	}
	sb.WriteString(fmt.Sprintf(" from `%s`.`%s` where ", tbl.OriginTable.Database, tbl.OriginTable.Table))
	sb.WriteString(tbl.Condition.String())

	return sb.String()
}

type ViewSingleCondition struct {
	Column Column
	Table  string
}

func (tbl *ViewSingleCondition) String() string {
	return fmt.Sprintf("`%s` = %q", tbl.Column.Name, tbl.Table)
}

type Row struct {
	Table          *Table
	AccountIdx     int
	Columns        []string
	RawColumns     []any
	Name2ColumnIdx map[string]int
}

func (tbl *Table) GetRow(ctx context.Context) *Row {
	row := &Row{
		Table:          tbl,
		AccountIdx:     -1,
		Columns:        make([]string, len(tbl.Columns)),
		RawColumns:     make([]any, len(tbl.Columns)),
		Name2ColumnIdx: make(map[string]int),
	}
	for idx, col := range tbl.Columns {
		row.Name2ColumnIdx[col.Name] = idx
	}
	if tbl.AccountColumn != nil {
		idx, exist := row.Name2ColumnIdx[tbl.AccountColumn.Name]
		if !exist {
			panic(moerr.NewInternalError(ctx, "%s table missing %s column", tbl.GetName(), tbl.AccountColumn.Name))
		}
		row.AccountIdx = idx
	}
	row.Reset()
	return row
}

func (r *Row) Reset() {
	for idx := 0; idx < len(r.Columns); idx++ {
		r.Columns[idx] = ""
	}
	for idx, typ := range r.Table.Columns {
		switch typ.ColType.ToType().Oid {
		case types.T_int64:
			r.RawColumns[idx] = int64(0)
		case types.T_uint64:
			r.RawColumns[idx] = uint64(0)
		case types.T_float64:
			r.RawColumns[idx] = float64(0)
		case types.T_char, types.T_varchar, types.T_blob, types.T_text:
			r.RawColumns[idx] = typ.Default
		case types.T_json:
			r.RawColumns[idx] = typ.Default
		case types.T_datetime:
			r.RawColumns[idx] = time.Time{}
		default:
			logutil.Errorf("the value type %v is not SUPPORT", typ.ColType.ToType().String())
			panic("the value type is not support now")
		}
	}
}

// GetAccount return r.Columns[r.AccountIdx] if r.AccountIdx >= 0 and r.Table.PathBuilder.SupportAccountStrategy,
// else return "sys"
func (r *Row) GetAccount() string {
	if r.Table.PathBuilder.SupportAccountStrategy() && r.AccountIdx >= 0 {
		return r.Columns[r.AccountIdx]
	}
	if r.Table.SupportConstAccess && len(r.Table.Account) > 0 {
		return r.Table.Account
	}
	return "sys"
}

func (r *Row) SetVal(col string, val string) {
	if idx, exist := r.Name2ColumnIdx[col]; !exist {
		logutil.Fatalf("column(%s) not exist in table(%s)", col, r.Table.Table)
	} else {
		r.Columns[idx] = val
	}
}

func (r *Row) SetColumnVal(col Column, val string) {
	if idx, exist := r.Name2ColumnIdx[col.Name]; !exist {
		logutil.Fatalf("column(%s) not exist in table(%s)", col.Name, r.Table.Table)
	} else {
		r.Columns[idx] = val
	}
}

func (r *Row) SetFloat64(col string, val float64) {
	r.SetVal(col, fmt.Sprintf("%f", val))
}

func (r *Row) SetInt64(col string, val int64) {
	r.SetVal(col, fmt.Sprintf("%d", val))
}

func (r *Row) SetRawColumnVal(col Column, val any) {
	if idx, exist := r.Name2ColumnIdx[col.Name]; !exist {
		logutil.Fatalf("column(%s) not exist in table(%s)", col.Name, r.Table.Table)
	} else {
		r.RawColumns[idx] = val
	}
}

// ToStrings output all column as string
func (r *Row) ToStrings() []string {
	for idx, col := range r.Columns {
		if len(col) == 0 {
			r.Columns[idx] = r.Table.Columns[idx].Default
		}
	}
	return r.Columns
}

func (r *Row) GetRawColumn() []any {
	return r.RawColumns
}

// ToRawStrings not format
func (r *Row) ToRawStrings() []string {
	return r.Columns
}

func (r *Row) ParseRow(cols []string) error {
	// fixme: check len(r.Name2ColumnIdx) != len(cols)
	r.Columns = cols
	return nil
}

func (r *Row) PrimaryKey() string {
	if len(r.Table.PrimaryKeyColumn) == 0 {
		return ""
	}
	if len(r.Table.PrimaryKeyColumn) == 1 {
		return r.Columns[r.Name2ColumnIdx[r.Table.PrimaryKeyColumn[0].Name]]
	}
	sb := strings.Builder{}
	for _, col := range r.Table.PrimaryKeyColumn {
		sb.WriteString(r.Columns[r.Name2ColumnIdx[col.Name]])
		sb.WriteRune('-')
	}
	return sb.String()
}

func (r *Row) Size() (size int64) {
	for _, col := range r.Columns {
		size += int64(len(col))
	}
	return
}

var _ TableOptions = (*NoopTableOptions)(nil)

type NoopTableOptions struct{}

func (o NoopTableOptions) FormatDdl(ddl string) string        { return ddl }
func (o NoopTableOptions) GetCreateOptions() string           { return "" }
func (o NoopTableOptions) GetTableOptions(PathBuilder) string { return "" }

var _ TableOptions = (*CsvTableOptions)(nil)

type CsvTableOptions struct {
	Formatter string
	DbName    string
	TblName   string
	Account   string
}

func getExternalTableDDLPrefix(sql string) string {
	return strings.Replace(sql, "CREATE TABLE", "CREATE EXTERNAL TABLE", 1)
}

func (o *CsvTableOptions) FormatDdl(ddl string) string {
	return getExternalTableDDLPrefix(ddl)
}

func (o *CsvTableOptions) GetCreateOptions() string {
	return "EXTERNAL "
}

func (o *CsvTableOptions) GetTableOptions(builder PathBuilder) string {
	if builder == nil {
		builder = NewDBTablePathBuilder()
	}
	if len(o.Formatter) > 0 {
		return fmt.Sprintf(o.Formatter, builder.BuildETLPath(o.DbName, o.TblName, o.Account))
	}
	return ""
}

func GetOptionFactory(ctx context.Context, engine string) func(db string, tbl string, account string) TableOptions {
	var infileFormatter = ` infile{"filepath"="etl:%s","compression"="none"}` +
		` FIELDS TERMINATED BY ',' ENCLOSED BY '"' LINES TERMINATED BY '\n' IGNORE 0 lines`
	switch engine {
	case NormalTableEngine:
		return func(_, _, _ string) TableOptions { return NoopTableOptions{} }
	case ExternalTableEngine:
		return func(db, tbl, account string) TableOptions {
			return &CsvTableOptions{Formatter: infileFormatter, DbName: db, TblName: tbl, Account: account}
		}
	default:
		panic(moerr.NewInternalError(ctx, "unknown engine: %s", engine))
	}
}

var gTable map[string]*Table
var mux sync.Mutex

// RegisterTableDefine return old one, if already registered
func RegisterTableDefine(table *Table) *Table {
	mux.Lock()
	defer mux.Unlock()
	if len(gTable) == 0 {
		gTable = make(map[string]*Table)
	}
	id := table.GetIdentify()
	old := gTable[id]
	gTable[id] = table
	return old
}

func GetAllTable() []*Table {
	mux.Lock()
	defer mux.Unlock()
	tables := make([]*Table, 0, len(gTable))
	for _, tbl := range gTable {
		tables = append(tables, tbl)
	}
	return tables
}

func GetTable(b string) (*Table, bool) {
	mux.Lock()
	defer mux.Unlock()
	tbl, exist := gTable[b]
	return tbl, exist
}

func SetPathBuilder(ctx context.Context, pathBuilder string) error {
	tables := GetAllTable()
	bp := PathBuilderFactory(pathBuilder)
	if bp == nil {
		return moerr.NewNotSupported(ctx, "not support PathBuilder: %s", pathBuilder)
	}
	for _, tbl := range tables {
		tbl.PathBuilder = bp
	}
	return nil
}
