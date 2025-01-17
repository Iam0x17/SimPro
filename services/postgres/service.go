package postgres

import (
	"SimPro/common"
	"SimPro/config"
	"context"
	"fmt"
	"github.com/auxten/postgresql-parser/pkg/sql/parser"
	wire "github.com/jeroenrinzema/psql-wire"
	"github.com/lib/pq/oid"
	"github.com/sirupsen/logrus"
	"log"
	"net"
	"sync"
)

var (
	postgresLogger *logrus.Logger
	accountsList   map[string]string
)

func init() {
	postgresLogger = common.SetupServiceLogger("Postgres", true)
}

type SimPostgresService struct {
	listener net.Listener
	wg       sync.WaitGroup
}

func (s *SimPostgresService) Stop() error {
	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			return fmt.Errorf("关闭监听器失败: %v", err)
		}
	}
	postgresLogger.Println("Postgres 服务已停止")
	return nil
}

// Serve方法处理FTP连接相关逻辑
func (s *SimPostgresService) Start(cfg *config.Config) error {

	accountsList = map[string]string{
		cfg.Postgres.User: cfg.Postgres.Pass,
	}

	var err error
	s.listener, err = net.Listen("tcp", ":"+cfg.Postgres.Port)
	if err != nil {
		return err
	}
	postgresLogger.Printf("Postgres 服务正在监听端口 %s", cfg.Postgres.Port)
	// 创建一个新的服务器实例，并设置认证策略
	server, err := wire.NewServer(
		handler,
		wire.SessionAuthStrategy(wire.ClearTextPassword(authenticate)),
	)
	if err != nil {
		fmt.Println("Error creating server:", err)
		return err
	}
	go server.Serve(s.listener)

	return nil
}

// GetServiceName方法返回服务名称
func (s *SimPostgresService) GetName() string {
	return "Postgres"
}

// 定义一个简单的认证函数
func authenticate(ctx context.Context, username, password string) (context.Context, bool, error) {
	// 在这里验证用户名和密码
	// 这个示例中只是简单地检查用户名和密码是否匹配
	//validUsers := map[string]string{
	//	"root":     "123456",
	//	"postgres": "123456",
	//}

	passwordHash, ok := accountsList[username]
	if !ok {
		return ctx, false, fmt.Errorf("invalid username")
	}

	if passwordHash != password {
		return ctx, false, fmt.Errorf("invalid password")
	}

	// 如果认证成功，返回一个包含用户名的上下文
	return context.WithValue(ctx, "username", username), true, nil
}

// 定义一个查询处理器

var table = wire.Columns{
	{
		Table: 0,
		Name:  "name",
		Oid:   oid.T_text,
		Width: 256,
	},
	{
		Table: 0,
		Name:  "member",
		Oid:   oid.T_bool,
		Width: 1,
	},
	{
		Table: 0,
		Name:  "age",
		Oid:   oid.T_int4,
		Width: 1,
	},
}

func handler(ctx context.Context, query string) (wire.PreparedStatements, error) {
	log.Println("incoming SQL query:", query)

	// handle := func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) error {
	// 	writer.Row([]any{"John", true, 29})
	// 	writer.Row([]any{"Marry", false, 21})
	// 	return writer.Complete("SELECT 2")
	// }

	// return wire.Prepared(wire.NewStatement(handle, wire.WithColumns(table))), nil
	// 使用 postgresql-parser 解析 SQL 查询
	stmt, err := parser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %v", err)
	}

	// 在这里，你可以对解析后的语句进行处理，例如打印或执行查询
	fmt.Println("Parsed query:", stmt)

	// 返回一个准备好的语句，这里我们只是简单地打印查询并返回一个成功状态
	return wire.Prepared(wire.NewStatement(func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) error {
		fmt.Println("Executing query:", query)
		return writer.Complete("OK")
	}, wire.WithColumns(table))), nil
}
