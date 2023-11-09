package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jtaylor-io/safe-as-houses/util"
	_ "github.com/lib/pq"
)

var (
	testQueries *Queries
	testDb      *sql.DB
)

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	testDb, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	testQueries = New(testDb)

	os.Exit(m.Run())
}
