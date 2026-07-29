package server

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/savsgio/atreugo/v11"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtTokenLifetime = 8 * time.Hour
	adminPasswordLen = 16
)

const adminPasswordAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// generateRandomPassword generates a cryptographically secure alphanumeric password.
func generateRandomPassword(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = adminPasswordAlphabet[int(b)%len(adminPasswordAlphabet)]
	}
	return string(out), nil
}

// issueJWT signs a new HS256 JWT for the given username using MetaData.Token as HMAC key.
func issueJWT(username string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(jwtTokenLifetime)
	claims := jwt.RegisteredClaims{
		Subject:   username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(MetaData.Token))
	if err != nil {
		return "", 0, err
	}
	return signed, exp.Unix(), nil
}

// authJWT validates the Authorization: Bearer <token> header and returns the parsed claims.
func authJWT(ctx *atreugo.RequestCtx) (jwt.Claims, bool) {
	raw := string(ctx.Request.Header.Peek("Authorization"))
	if raw == "" || !strings.HasPrefix(raw, "Bearer ") {
		return nil, false
	}
	tokenStr := strings.TrimPrefix(raw, "Bearer ")
	claims := &jwt.RegisteredClaims{}
	tok, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(MetaData.Token), nil
	})
	if err != nil || !tok.Valid {
		return nil, false
	}
	return claims, true
}

// LoginHandler authenticates the admin account and returns a JWT.
func LoginHandler(ctx *atreugo.RequestCtx) error {
	var req LoginRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		return ctx.JSONResponse(FailedWithS("invalid credentials", 401), 401)
	}

	admin := MetaData.Admin
	if admin == nil || admin.UserName == "" || admin.Password == "" {
		return ctx.JSONResponse(FailedWithS("invalid credentials", 401), 401)
	}

	if req.Username != admin.UserName {
		return ctx.JSONResponse(FailedWithS("invalid credentials", 401), 401)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return ctx.JSONResponse(FailedWithS("invalid credentials", 401), 401)
	}

	token, expiresAt, err := issueJWT(admin.UserName)
	if err != nil {
		return ctx.JSONResponse(Failed("failed to issue token"), 500)
	}

	return ctx.JSONResponse(SuccessWithD(LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}))
}
