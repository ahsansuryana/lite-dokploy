package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Manager struct {
	jwtSecret    []byte
	username     string
	passwordHash string
}

type Claims struct {
	Username string `json:"username"`
	Exp      int64  `json:"exp"`
}

func NewManager(username, password, jwtSecret string) (*Manager, error) {
	if jwtSecret == "" {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return nil, fmt.Errorf("generate jwt secret: %w", err)
		}
		jwtSecret = base64.RawURLEncoding.EncodeToString(buf)
	}

	hash := ""
	if password != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		hash = string(h)
	}

	return &Manager{
		jwtSecret:    []byte(jwtSecret),
		username:     username,
		passwordHash: hash,
	}, nil
}

func (m *Manager) Enabled() bool {
	return m.passwordHash != ""
}

func (m *Manager) Authenticate(username, password string) bool {
	if !m.Enabled() {
		return true
	}
	if username != m.username {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(m.passwordHash), []byte(password)) == nil
}

func (m *Manager) GenerateToken(username string) (string, error) {
	claims := Claims{
		Username: username,
		Exp:      time.Now().Add(24 * time.Hour).Unix(),
	}
	payload, _ := json.Marshal(claims)
	encoded := base64.RawURLEncoding.EncodeToString(payload)

	mac := hmac.New(sha256.New, m.jwtSecret)
	mac.Write([]byte(encoded))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return encoded + "." + sig, nil
}

func (m *Manager) ValidateToken(token string) (*Claims, error) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid token format")
	}

	mac := hmac.New(sha256.New, m.jwtSecret)
	mac.Write([]byte(parts[0]))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid signature")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal claims: %w", err)
	}

	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}
