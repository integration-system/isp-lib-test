package postgres

import (
	"github.com/go-pg/pg"
	"github.com/integration-system/isp-lib-test/utils"
	"github.com/integration-system/isp-lib/database"
	"github.com/integration-system/isp-lib/structure"
	"time"
)

func Wait(dbConfiguration structure.DBConfiguration, timeout time.Duration) (*pg.DB, error) {
	conn, err := utils.AwaitConnection(func() (interface{}, error) {
		c, err := database.NewDbConnection(dbConfiguration)
		return c, err
	}, timeout)
	if err != nil {
		return nil, err
	}
	return conn.(*pg.DB), nil
}
