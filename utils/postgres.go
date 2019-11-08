package utils

import (
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib/database"
	"github.com/integration-system/isp-lib/structure"
	"time"
)

func WaitPostgres(dbConfiguration structure.DBConfiguration, timeout time.Duration) (*pg.DB, error) {
	conn, err := AwaitConnection(func() (interface{}, error) {
		c, err := database.NewDbConnection(dbConfiguration)
		return c, err
	}, timeout)
	if err != nil {
		return nil, err
	}
	return conn.(*pg.DB), nil
}
