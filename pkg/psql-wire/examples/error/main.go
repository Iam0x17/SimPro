package main

import (
	"context"
	"errors"
	"log"

	wire "SimPro/pkg/psql-wire"
	"SimPro/pkg/psql-wire/codes"
	psqlerr "SimPro/pkg/psql-wire/errors"
)

func main() {
	log.Println("PostgreSQL server is up and running at [127.0.0.1:5432]")
	wire.ListenAndServe("127.0.0.1:5432", handler)
}

func handler(ctx context.Context, query string) (wire.PreparedStatements, error) {
	log.Println("incoming SQL query:", query)

	err := errors.New("unimplemented feature")
	err = psqlerr.WithCode(err, codes.FeatureNotSupported)
	err = psqlerr.WithSeverity(err, psqlerr.LevelFatal)

	return nil, err
}
