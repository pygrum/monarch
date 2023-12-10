package teamserver

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/db"
	"github.com/pygrum/monarch/pkg/teamserver/roles"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct{}
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func newWrappedStream(ss grpc.ServerStream, ctx context.Context) *wrappedStream {
	return &wrappedStream{ss, ctx}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, err := a.authorize(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx, err := a.authorize(ss.Context(), info.FullMethod)
		if err != nil {
			return err
		}
		return handler(srv, newWrappedStream(ss, ctx))
	}
}

func (a *AuthInterceptor) authorize(ctx context.Context, method string) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, errors.New("no auth metadata provided")
	}
	v1 := md["uid"]
	v2 := md["challenge"]
	if len(v1) == 0 || len(v2) == 0 {
		return ctx, errors.New("authorization info not provided")
	}
	uid := v1[0]
	challenge := v2[0]
	player := &db.Player{}
	if err := db.FindOneConditional("uuid = ?", uid, &player); err != nil {
		return ctx, fmt.Errorf("query failed: %v", err)
	}
	if len(player.UUID) == 0 {
		return ctx, errors.New("this player doesn't exist or was kicked from the server")
	}
	secret, _ := hex.DecodeString(player.Secret)
	plainChallenge, err := crypto.DecryptAES(secret, challenge)
	if err != nil {
		return ctx, fmt.Errorf("challenge decryption failed: %v", err)
	}
	if plainChallenge != player.Challenge {
		return ctx, fmt.Errorf("challenge string doesn't match - player is unauthorized")
	}
	if !roles.AuthorizedEndpoint(player.Role, method) {
		return ctx, errors.New("player is unauthorized")
	}
	md["role"] = []string{string(player.Role)}
	md["player"] = []string{player.UUID}
	ctx = metadata.NewIncomingContext(ctx, md)

	return ctx, nil
}

func NewAuthInterceptor() *AuthInterceptor {
	return &AuthInterceptor{}
}
