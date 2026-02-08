package database

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestConnectDbWithoutDsn(t *testing.T) {
	a := assert.New(t)
	err := ConnectDb("")
	a.NotNil(err)
}

func TestConnectDbWithDsn(t *testing.T) {
	a := assert.New(t)
	_ = godotenv.Load("../../.env")
	dsn, dsnErr := os.LookupEnv("DATABASE_DSN")
	a.True(dsnErr)
	err := ConnectDb(dsn)
	a.Nil(err)
	a.NotNil(Pool)
	a.IsType(&pgxpool.Pool{}, Pool)
	Pool.Close()
	ctx := context.Background()
	err = Pool.Ping(ctx)
	a.NotNil(err)
}
