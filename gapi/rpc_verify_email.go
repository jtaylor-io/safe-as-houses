package gapi

import (
	"context"

	db "github.com/jtaylor-io/safe-as-houses/db/sqlc"
	"github.com/jtaylor-io/safe-as-houses/pb"
	"github.com/jtaylor-io/safe-as-houses/val"
	"github.com/rs/zerolog/log"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) VerifyEmail(
	ctx context.Context,
	req *pb.VerifyEmailRequest,
) (*pb.VerifyEmailResponse, error) {
	txParams := db.VerifyEmailTxParams{
		EmailId:    req.GetEmailId(),
		SecretCode: req.GetSecretCode(),
	}
	log.Info().Msgf("EmailId: %d, SecretCode: %s", txParams.EmailId, txParams.SecretCode)
	txResult, err := server.store.VerifyEmailTx(ctx, txParams)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email")
	}
	rsp := &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}
	return rsp, nil
}

func validateVerifyEmailRequest(
	req *pb.VerifyEmailRequest,
) []*errdetails.BadRequest_FieldViolation {
	var violations []*errdetails.BadRequest_FieldViolation

	if err := val.ValidateEmailId(req.GetEmailId()); err != nil {
		violations = append(violations, fieldViolation("email_id", err))
	}
	if err := val.ValidateSecretCode(req.GetSecretCode()); err != nil {
		violations = append(violations, fieldViolation("secret_code", err))
	}
	return violations
}
