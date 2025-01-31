package api

import (
	"fmt"

	"github.com/nhat195/simple_bank/token"
	"github.com/nhat195/simple_bank/util"
	"github.com/nhat195/simple_bank/worker"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	db "github.com/nhat195/simple_bank/db/sqlc"
)

// Server serves HTTP requests for our banking service.
type Server struct {
	config          util.Config
	store           db.Store
	router          *gin.Engine
	tokenMaker      token.Maker
	taskDistributor worker.TaskDistributor
}

// NewServer creates a new HTTP server and set up routing.
func NewServer(config util.Config, store db.Store) (*Server, error) {

	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker %w", err)
	}

	server := &Server{
		store:      store,
		config:     config,
		tokenMaker: tokenMaker,
		// taskDistributor: taskDistributor,
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validatorCurrency)
	}

	server.setupRouter()

	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", server.createUser)
	router.POST("/users/login", server.loginUser)

	router.POST("/token/renew_access", server.renewAccessToken)

	authGroup := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authGroup.POST("/accounts", server.createAccount)

	authGroup.GET("/accounts/:id", server.getAccount)

	authGroup.GET("/accounts", server.listAccounts)

	authGroup.POST("/transfer", server.createTransfer)

	server.router = router
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
