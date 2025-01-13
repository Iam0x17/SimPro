package mysql

import (
	"SimPro/common"
	"context"
	"net"
	"time"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/sirupsen/logrus"
)

var mySqlLogger *logrus.Logger

var (
	dbName    = "ZC"
	tableName = "mytable"
	address   = "localhost"
	port      = 3306
)

func init() {
	mySqlLogger = common.SetupServiceLogger("Mysql", true)
}

// SimMysqlService 实现通用的MockService接口
type SimMySqlService struct{}

func (m *SimMySqlService) NeedsListener() bool {
	return true
}

func (m *SimMySqlService) Serve(ctx context.Context, conn net.Conn) {

}

// Serve方法处理FTP连接相关逻辑
func (m *SimMySqlService) ServeWithListener(ctx context.Context, listener net.Listener) {
	pro := createTestDatabase()
	engine := sqle.NewDefault(pro)

	session := memory.NewSession(sql.NewBaseSession(), pro)
	ctxSql := sql.NewContext(context.Background(), sql.WithSession(session))
	ctxSql.SetCurrentDatabase(dbName)

	mysqlDb := engine.Analyzer.Catalog.MySQLDb

	go func() {
		ed := mysqlDb.Editor()
		defer ed.Close()
		mysqlDb.AddSuperUser(ed, "root", "localhost", "zczc")
	}()

	config := server.Config{
		Listener: listener,
	}
	s, err := server.NewServer(config, engine, memory.NewSessionBuilder(pro), nil)
	if err != nil {
		panic(err)
	}
	if err = s.Start(); err != nil {
		panic(err)
	}
}

// GetServiceName方法返回服务名称
func (m *SimMySqlService) GetServiceName() string {
	return "MySql"
}

// For go-mysql-server developers: Remember to update the snippet in the README when this file changes.

func createTestDatabase() *memory.DbProvider {
	db := memory.NewDatabase(dbName)
	db.BaseDatabase.EnablePrimaryKeyIndexes()

	pro := memory.NewDBProvider(db)
	session := memory.NewSession(sql.NewBaseSession(), pro)
	ctx := sql.NewContext(context.Background(), sql.WithSession(session))

	table := memory.NewTable(db, tableName, sql.NewPrimaryKeySchema(sql.Schema{
		{Name: "name", Type: types.Text, Nullable: false, Source: tableName, PrimaryKey: true},
		{Name: "email", Type: types.Text, Nullable: false, Source: tableName, PrimaryKey: true},
		{Name: "phone_numbers", Type: types.JSON, Nullable: false, Source: tableName},
		{Name: "created_at", Type: types.MustCreateDatetimeType(query.Type_DATETIME, 6), Nullable: false, Source: tableName},
	}), db.GetForeignKeyCollection())
	db.AddTable(tableName, table)

	creationTime := time.Unix(0, 1667304000000001000).UTC()
	_ = table.Insert(ctx, sql.NewRow("Jane Deo", "janedeo@gmail.com", types.MustJSON(`["556-565-566", "777-777-777"]`), creationTime))
	_ = table.Insert(ctx, sql.NewRow("Jane Doe", "jane@doe.com", types.MustJSON(`[]`), creationTime))
	_ = table.Insert(ctx, sql.NewRow("John Doe", "john@doe.com", types.MustJSON(`["555-555-555"]`), creationTime))
	_ = table.Insert(ctx, sql.NewRow("John Doe", "johnalt@doe.com", types.MustJSON(`[]`), creationTime))

	return pro
}
