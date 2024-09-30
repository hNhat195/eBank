package token

import "time"

type Maker interface {
	CreateToken(username string, duration time.Duration) (string, *Payload, error)

	// VerifyToken parses a token string and validates the payload
	VerifyToken(token string) (*Payload, error)
}
