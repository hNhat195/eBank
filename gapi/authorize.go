package gapi

import (
	"context"
	"fmt"
	"strings"

	"github.com/nhat195/simple_bank/token"
	"google.golang.org/grpc/metadata"
)

const (
	authorizationHeader = "authorization"
	authorizationBearer = "bearer"
)

func (server *Server) authorizeUser(ctx context.Context) (*token.Payload, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("metadata is not provided")
	}
	value := md.Get(authorizationHeader)
	if len(value) == 0 {
		return nil, fmt.Errorf("authorization token is not provided")
	}

	authHeader := value[0]
	field := strings.Fields(authHeader)
	if len(field) < 2 || strings.ToLower(field[0]) != authorizationBearer {
		return nil, fmt.Errorf("authorization token is not provided in bearer format")
	}

	accessToken := field[1]
	payload, err := server.tokenMaker.VerifyToken(accessToken)

	if err != nil {
		return nil, fmt.Errorf("access token is invalid: %v", err)
	}

	return payload, nil

}
