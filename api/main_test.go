package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	db "github.com/nhat195/simple_bank/db/sqlc"
	"github.com/nhat195/simple_bank/util"
)

func NewTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store)
	if err != nil {
		t.Fatalf("cannot create server: %v", err)
	}

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}
