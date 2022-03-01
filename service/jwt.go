package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/zhupanovdm/gophermart/config"
	"github.com/zhupanovdm/gophermart/model/user"
	"github.com/zhupanovdm/gophermart/pkg/logging"
)

const (
	jwtServiceName = "JWT Service"

	userIDClaim     = "usr"
	expirationClaim = "exp"
)

type jwtImpl struct {
	secret []byte
	ttl    time.Duration
}

func (j *jwtImpl) Token(ctx context.Context, usr *user.User) (user.Token, error) {
	ctx, logger := logging.ServiceLogger(ctx, jwtServiceName, logging.With(usr))
	logger.Info().Msg("retrieving security token")

	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		userIDClaim:     usr.ID,
		expirationClaim: time.Now().Add(j.ttl).Unix(),
	}).SignedString(j.secret)

	if err != nil {
		logger.Err(err).Msg("failed to create token")
		return user.VoidToken, err
	}

	logger.Trace().Msg("token created")
	return user.Token(t), nil
}

func (j *jwtImpl) Authenticate(ctx context.Context, authToken user.Token) (user.ID, error) {
	ctx, logger := logging.ServiceLogger(ctx, jwtServiceName)
	logger.Info().Msg("authenticating with token")

	token, err := jwt.Parse(string(authToken), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if f, ok := claims[userIDClaim].(float64); ok {
			userID := user.ID(f)
			logger.UpdateContext(logging.ContextWith(userID))
			logger.Trace().Msg("token valid")
			return userID, nil
		}
		logger.Warn().Msg("no user id claim")
		return user.VoidID, errors.New("verification failed")
	}
	return user.VoidID, err
}

func NewJWT(cfg *config.Config) JWT {
	return &jwtImpl{
		secret: []byte(cfg.JWTSecret),
		ttl:    cfg.JWTTTL,
	}
}
