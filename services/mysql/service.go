package mysql

import (
	"SimPro/common"
	"SimPro/config"
	"context"
	"fmt"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"

	sqle "SimPro/pkg/go-mysql-server"
	"SimPro/pkg/go-mysql-server/memory"
	"SimPro/pkg/go-mysql-server/server"
	"SimPro/pkg/go-mysql-server/sql"
	"SimPro/pkg/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"
)

var (
	dbName    = "ZC"
	tableName = "mytable"
)

// SimMysqlService 实现通用的MockService接口
type SimMySqlService struct {
	listener net.Listener
	wg       sync.WaitGroup
	running  bool
}

func (s *SimMySqlService) Stop() error {
	if s.listener != nil {
		s.running = false
		err := s.listener.Close()
		if err != nil {
			return err
		}
	}
	common.Logger.Info(common.EventStopService, zap.String("protocol", "mysql"), zap.String("info", "MySql service has stopped"))
	//mySqlLogger.Println("MySql 服务已停止")
	return nil
}

// Serve方法处理FTP连接相关逻辑
func (s *SimMySqlService) Start(cfg *config.Config) error {

	var err error
	s.listener, err = net.Listen("tcp", ":"+cfg.MySql.Port)
	if err != nil {
		return err
	}

	common.Logger.Info(common.EventStartService, zap.String("protocol", "mysql"), zap.String("info", fmt.Sprintf("MySql service is listening on port %s", cfg.MySql.Port)))
	pro := createTestDatabase()
	engine := sqle.NewDefault(pro)

	session := memory.NewSession(sql.NewBaseSession(), pro)
	ctxSql := sql.NewContext(context.Background(), sql.WithSession(session))
	ctxSql.SetCurrentDatabase(dbName)

	mysqlDb := engine.Analyzer.Catalog.MySQLDb

	go func() {
		ed := mysqlDb.Editor()
		defer ed.Close()
		mysqlDb.AddSuperUser(ed, cfg.MySql.User, "localhost", cfg.MySql.Pass)
	}()

	cfgServer := server.Config{
		Listener: s.listener,
	}
	sServer, err := server.NewServer(cfgServer, engine, memory.NewSessionBuilder(pro), nil)
	if err != nil {
		panic(err)
	}
	//if err = sServer.Start(); err != nil {
	//	panic(err)
	//}
	//err = sServer.Close()
	//if err != nil {
	//	return err
	//}
	go sServer.Start()
	return nil
}

// GetServiceName方法返回服务名称
func (m *SimMySqlService) GetName() string {
	return "mysql"
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
