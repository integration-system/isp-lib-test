package postgres

import (
	"time"

	"github.com/go-pg/pg/v9"
	"github.com/integration-system/isp-lib-test/utils"
	"github.com/integration-system/isp-lib/v2/database"
	"github.com/integration-system/isp-lib/v2/structure"
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
