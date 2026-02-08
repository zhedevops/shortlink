package database

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestConnectDBWithoutDsn(t *testing.T) {
	a := assert.New(t)
	err := ConnectDB("")
	a.NotNil(err)
}

func TestConnectDBWithDsn(t *testing.T) {
	a := assert.New(t)
	_ = godotenv.Load("../../.env")
	dsn, dsnErr := os.LookupEnv("DATABASE_DSN")
	a.True(dsnErr)
	err := ConnectDB(dsn)
	a.Nil(err)
	a.NotNil(Pool)
	a.IsType(&pgxpool.Pool{}, Pool)
	CloseDB()
	ctx := context.Background()
	err = Pool.Ping(ctx)
	a.NotNil(err)
}
