
-- test system db table_number
-- @bvt:issue#7564
show table_number from mo_task;
show table_number from information_schema;
show table_number from mysql;
show table_number from mo_catalog;
show table_number from system_metrics;
show table_number from system;
-- @bvt:issue

-- test system tables column_number
-- @bvt:issue#7564
use mo_task;
show column_number from sys_async_task;
show column_number from sys_cron_task;

use information_schema;
show column_number from key_column_usage;
show column_number from columns;
show column_number from profiling;
-- @bvt:issue

-- @bvt:issue#7520
show column_number from PROCESSLIST;
-- @bvt:issue

-- @bvt:issue#7564
show column_number from schemata;
show column_number from character_sets;
show column_number from triggers;
show column_number from tables;
show column_number from engines;
show column_number from routines;
show column_number from parameters;
show column_number from keywords;

use mysql;
show column_number from user;
show column_number from db;
show column_number from procs_priv;
show column_number from columns_priv;
show column_number from tables_priv;

use mo_catalog;
show column_number from mo_user;
show column_number from mo_account;
show column_number from mo_role;
show column_number from mo_user_grant;
show column_number from mo_role_grant;
show column_number from mo_role_privs;
show column_number from mo_user_defined_function;
show column_number from mo_tables;
show column_number from mo_database;
show column_number from mo_columns;

use system_metrics;
show column_number from metric;
show column_number from sql_statement_total;
show column_number from sql_statement_errors;
show column_number from sql_transaction_total;
show column_number from sql_transaction_errors;
show column_number from server_connections;
show column_number from process_cpu_percent;
show column_number from process_resident_memory_bytes;
show column_number from process_open_fds;
show column_number from sys_cpu_seconds_total;
show column_number from sys_cpu_combined_percent;
show column_number from sys_memory_used;
show column_number from sys_memory_available;
show column_number from sys_disk_read_bytes;
show column_number from sys_disk_write_bytes;
show column_number from sys_net_recv_bytes;
show column_number from sys_net_sent_bytes;

use system;
show column_number from statement_info;
show column_number from rawlog;
show column_number from log_info;
show column_number from error_info;
show column_number from span_info;
-- @bvt:issue

-- test max nad min values of the data in the table
drop database if exists test_db;
create database test_db;
-- @bvt:issue#7564
show table_number from test_db;
-- @bvt:issue
use test_db;

drop table if exists t1;
-- test non primary key table
create table t1(
col1 int,
col2 float,
col3 varchar,
col4 blob,
col6 date,
col7 bool
);

-- @bvt:issue#7564
show table_number from test_db;
-- @bvt:issue

show table_values from t1;
select mo_table_rows("test_db","t1"),mo_table_size("test_db","t1");


insert into t1 values(100,10.34,"你好",'aaa','2011-10-10',0);
show table_values from t1;

insert into t1 values(10,1.34,"你",'aa','2011-10-11',1);
show table_values from t1;

select mo_table_rows("test_db","t1"),mo_table_size("test_db","t1");

-- test primary key table
drop table if exists t11;
create table t11(
col1 int primary key,
col2 float,
col3 varchar,
col4 blob,
col6 date,
col7 bool
);

-- @bvt:issue#7564
show table_number from test_db;
-- @bvt:issue

show table_values from t11;
select mo_table_rows("test_db","t11"),mo_table_size("test_db","t11");

insert into t11 values(100,10.34,"你好",'aaa','2011-10-10',0);
show table_values from t11;

insert into t11 values(10,1.34,"你",'aa','2011-10-11',1);
show table_values from t11;

select mo_table_rows("test_db","t11"),mo_table_size("test_db","t11");

-- test external table
create external table external_table(
col1 int,
col2 float,
col3 varchar,
col4 blob,
col6 date,
col7 bool
)infile{"filepath"='$resources/external_table_file/external_table.csv'} fields terminated by ',' enclosed by '\"' lines terminated by '\n';

select * from external_table;
-- @bvt:issue#7564
show table_number from test_db;
-- @bvt:issue

-- @bvt:issue#7568
show table_values from external_table;
-- @bvt:issue


create table t2(
col1 json
);

-- @bvt:issue#7527
show table_values from t2;

insert into t2 values();
show table_values from t2;

insert into t2 values(('{"x": 17}'));
show table_values from t2;

insert into t2 values (('{"x": [18]}'));
show table_values from t2;
-- @bvt:issue

create table t3(
col1 decimal
);

-- @bvt:issue#7562
show table_values from t3;

insert into t3 values();
show table_values from t3;

insert into t3 values(3.3);
show table_values from t3;

insert into t3 values(3.2);
show table_values from t3;
-- @bvt:issue
drop database test_db;


-- test common tenant system db table_number
drop account if exists test_account;
create account test_account admin_name = 'test_user' identified by '111';
-- @session:id=2&user=test_account:test_user&password=111
-- @bvt:issue#7564
show table_number from information_schema;
show table_number from mysql;
show table_number from mo_catalog;
show table_number from system_metrics;
show table_number from system;

use information_schema;
show column_number from key_column_usage;
show column_number from columns;
show column_number from profiling;
-- @bvt:issue#7564

-- @bvt:issue#7520
show column_number from PROCESSLIST;
-- @bvt:issue
-- @bvt:issue#7564
show column_number from schemata;
show column_number from character_sets;
show column_number from triggers;
show column_number from tables;
show column_number from engines;
show column_number from routines;
show column_number from parameters;
show column_number from keywords;

use mysql;
show column_number from user;
show column_number from db;
show column_number from procs_priv;
show column_number from columns_priv;
show column_number from tables_priv;

use mo_catalog;
show column_number from mo_user;
show column_number from mo_role;
show column_number from mo_user_grant;
show column_number from mo_role_grant;
show column_number from mo_role_privs;
show column_number from mo_user_defined_function;
show column_number from mo_tables;
show column_number from mo_database;
show column_number from mo_columns;

use system;
show column_number from statement_info;
-- @bvt:issue

-- test max nad min values of the data in the table
drop database if exists test_db;
create database test_db;

-- @bvt:issue#7564
show table_number from test_db;
-- @bvt:issue

use test_db;

drop table if exists t1;
-- test non primary key table
create table t1(
col1 int,
col2 float,
col3 varchar,
col4 blob,
col6 date,
col7 bool
);

-- @bvt:issue#7564
show table_number from test_db;
-- @bvt:issue

show table_values from t1;
select mo_table_rows("test_db","t1"),mo_table_size("test_db","t1");


insert into t1 values(100,10.34,"你好",'aaa','2011-10-10',0);
show table_values from t1;

insert into t1 values(10,1.34,"你",'aa','2011-10-11',1);
show table_values from t1;

select mo_table_rows("test_db","t1"),mo_table_size("test_db","t1");

-- test primary key table
drop table if exists t11;
create table t11(
col1 int primary key,
col2 float,
col3 varchar,
col4 blob,
col6 date,
col7 bool
);

-- @bvt:issue#7564
show table_number from test_db;
-- @bvt:issue

show table_values from t11;
select mo_table_rows("test_db","t11"),mo_table_size("test_db","t11");

insert into t11 values(100,10.34,"你好",'aaa','2011-10-10',0);
show table_values from t11;

insert into t11 values(10,1.34,"你",'aa','2011-10-11',1);
show table_values from t11;

select mo_table_rows("test_db","t11"),mo_table_size("test_db","t11");

-- test external table
create external table external_table(
col1 int,
col2 float,
col3 varchar,
col4 blob,
col6 date,
col7 bool
)infile{"filepath"='$resources/external_table_file/external_table.csv'} fields terminated by ',' enclosed by '\"' lines terminated by '\n';

select * from external_table;
-- @bvt:issue#7564
show table_number from test_db;
-- @bvt:issue

-- @bvt:issue#7568
show table_values from external_table;
-- @bvt:issue


create table t2(
col1 json
);

-- @bvt:issue#7527
show table_values from t2;

insert into t2 values();
show table_values from t2;

insert into t2 values(('{"x": 17}'));
show table_values from t2;

insert into t2 values (('{"x": [18]}'));
show table_values from t2;
-- @bvt:issue

create table t3(
col1 decimal
);

-- @bvt:issue#7562
show table_values from t3;

insert into t3 values();
show table_values from t3;

insert into t3 values(3.3);
show table_values from t3;

insert into t3 values(3.2);
show table_values from t3;
-- @bvt:issue
drop database test_db;
-- @session


drop account if exists test_account;
