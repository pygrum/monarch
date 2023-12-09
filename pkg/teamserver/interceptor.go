package teamserver

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/pygrum/monarch/pkg/crypto"
	"github.com/pygrum/monarch/pkg/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthInterceptor struct{}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		err := a.authorize(ctx)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

func (a *AuthInterceptor) authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.New("no auth metadata provided")
	}
	v1 := md["uid"]
	v2 := md["challenge"]
	if len(v1) == 0 || len(v2) == 0 {
		return errors.New("authorization info not provided")
	}
	uid := v1[0]
	challenge := v2[0]
	player := &db.Player{}
	if err := db.FindOneConditional("uuid = ?", uid, &player); err != nil {
		return fmt.Errorf("query failed: %v", err)
	}
	if len(player.UUID) == 0 {
		return errors.New("this player doesn't exist or was kicked from the server")
	}
	secret, _ := hex.DecodeString(player.Secret)
	plainChallenge, err := crypto.DecryptAES(secret, challenge)
	if err != nil {
		return fmt.Errorf("challenge decryption failed: %v", err)
	}
	if plainChallenge != player.Challenge {
		return fmt.Errorf("challenge string doesn't match - player is unauthorized")
	}
	return nil
}

func NewAuthInterceptor() *AuthInterceptor {
	return &AuthInterceptor{}
}
