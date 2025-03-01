package driver

import (
	"context"

	"SimPro/pkg/go-mysql-server/memory"
	"SimPro/pkg/go-mysql-server/sql"
)

// A SessionBuilder creates SQL sessions.
type SessionBuilder interface {
	NewSession(ctx context.Context, id uint32, conn *Connector) (sql.Session, error)
}

// DefaultSessionBuilder creates basic SQL sessions.
type DefaultSessionBuilder struct {
	provider sql.DatabaseProvider
}

func NewDefaultSessionBuilder(provider sql.DatabaseProvider) *DefaultSessionBuilder {
	return &DefaultSessionBuilder{
		provider: provider,
	}
}

// NewSession calls sql.NewBaseSessionWithClientServer.
func (d DefaultSessionBuilder) NewSession(ctx context.Context, id uint32, conn *Connector) (sql.Session, error) {
	return memory.NewSession(sql.NewBaseSession(), d.provider), nil
}
