package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/bloomyindev/time-tracker/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const sessionTTL = 24 * time.Hour

type Claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

type Service struct {
	db       *sql.DB
	secret   []byte
	sessions *SessionStore
}

func NewService(db *sql.DB, secret string) *Service {
	return &Service{db: db, secret: []byte(secret), sessions: NewSessionStore(sessionTTL)}
}

func (s *Service) Register(email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`INSERT INTO users (email, password_hash) VALUES (?, ?)`, email, string(hash))
	return err
}

func (s *Service) authenticate(email, password string) (models.User, error) {
	var user models.User
	err := s.db.QueryRow(`SELECT id, email, password_hash FROM users WHERE email = ?`, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return models.User{}, ErrInvalidCredentials
	}
	if err != nil {
		return models.User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return models.User{}, ErrInvalidCredentials
	}
	return user, nil
}

// Login is used by the SSR cookie flow: it returns an opaque, revocable
// session token backed by the in-memory SessionStore.
func (s *Service) Login(email, password string) (string, error) {
	user, err := s.authenticate(email, password)
	if err != nil {
		return "", err
	}
	return s.sessions.Create(user.ID)
}

func (s *Service) Logout(token string) {
	s.sessions.Delete(token)
}

// IssueAPIToken mints a signed JWT for the given credentials. Kept for a
// future bearer-token API flow; not used by the SSR cookie login.
func (s *Service) IssueAPIToken(email, password string) (string, error) {
	user, err := s.authenticate(email, password)
	if err != nil {
		return "", err
	}

	claims := Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Email,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// VerifyAPIToken validates a JWT issued by IssueAPIToken. Kept alongside
// IssueAPIToken for the future API flow.
func (s *Service) VerifyAPIToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
