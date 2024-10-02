package gapi

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	db "github.com/nhat195/simple_bank/db/sqlc"
	"github.com/nhat195/simple_bank/pb"
	"github.com/nhat195/simple_bank/util"
	"github.com/nhat195/simple_bank/val"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (sever *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	authPayload, err := sever.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}
	if authPayload.Username != req.GetUsername() {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's profile")
	}

	violations := validateUpdateUserRequest(req)
	if len(violations) > 0 {
		return nil, invalidArgumentError(violations)
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		FullName: sql.NullString{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
		Email: sql.NullString{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
	}

	if req.Password != nil {
		hassPasswrord, err := util.HashPassword(req.GetPassword())

		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to hash password: %v", err)
		}
		arg.HashedPassword = sql.NullString{
			String: hassPasswrord,
			Valid:  true,
		}
		arg.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	acc, err := sever.store.UpdateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "no_data_found":
				return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
			}

		}
		return nil, status.Errorf(codes.Internal, "Failed to update user: %v", err)
	}
	rsp := &pb.UpdateUserResponse{
		User: convertUser(acc),
	}
	return rsp, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if req.Password != nil {
		if err := val.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}

	if req.FullName != nil {
		if err := val.ValidateFullName(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	if req.Email != nil {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	return violations
}
