package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"sourcecraft.dev/lct-hackaton-2026.arena/internal/entity"
)

type userRepository interface {
	CreateUser(ctx context.Context, user entity.User) (*entity.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)
	AddUserRole(ctx context.Context, userID string, role entity.ProductRole) error
	GetUserRoles(ctx context.Context, userID string) ([]entity.ProductRole, error)
}

type authRepository interface {
	CreateRefreshToken(ctx context.Context, token entity.RefreshToken) (*entity.RefreshToken, error)
	GetRefreshTokenByHashForUpdate(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	RotateRefreshToken(ctx context.Context, id, currentTokenHash, newTokenHash string, expiresAt time.Time) (*entity.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id string) error
}

type transactor interface {
	WithTx(ctx context.Context, f func(ctx context.Context) error) error
}

type authService struct {
	userRepository userRepository
	authRepository authRepository
	transactor     transactor
	cfg            *Config
}

type Config struct {
	JWTSecret       []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type accessTokenClaims struct {
	UserID     string `json:"user_id"`
	ActiveRole string `json:"active_role"`
	SessionID  string `json:"session_id"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepository userRepository, authRepository authRepository, transactor transactor, cfg Config) *authService {
	return &authService{
		userRepository: userRepository,
		authRepository: authRepository,
		transactor:     transactor,
		cfg:            &cfg,
	}
}

func (s *authService) Register(ctx context.Context, params entity.RegisterUser) (*entity.JWT, error) {
	passwordHash, err := hashPassword(params.Password)
	if err != nil {
		return nil, err
	}
	userID, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}

	var tokens *entity.JWT
	err = s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		createdUser, err := s.userRepository.CreateUser(txCtx, entity.User{
			ID:           userID,
			Username:     params.Username,
			Email:        params.Email,
			PasswordHash: passwordHash,
		})
		if err != nil {
			return err
		}
		if err := s.userRepository.AddUserRole(txCtx, createdUser.ID, entity.RoleLearner); err != nil {
			return err
		}
		tokens, err = s.createTokenPair(txCtx, createdUser.ID, entity.RoleLearner)
		return err
	})
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *authService) Login(ctx context.Context, params entity.LoginUser) (*entity.JWT, error) {
	user, err := s.userRepository.GetUserByEmail(ctx, params.Email)
	if errors.Is(err, entity.ErrUserNotFound) {
		return nil, entity.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if !verifyPassword(user.PasswordHash, params.Password) {
		return nil, entity.ErrInvalidCredentials
	}
	roles, err := s.userRepository.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	activeRole := entity.RoleLearner
	if !roleInList(activeRole, roles) && len(roles) > 0 {
		activeRole = roles[0]
	}
	return s.createTokenPair(ctx, user.ID, activeRole)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*entity.JWT, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, entity.ErrRefreshTokenMissing
	}
	var tokens *entity.JWT
	err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		currentTokenHash := hashRefreshToken(refreshToken)
		currentToken, err := s.authRepository.GetRefreshTokenByHashForUpdate(txCtx, currentTokenHash)
		if err != nil {
			return err
		}
		if err := validateRefreshTokenState(currentToken); err != nil {
			return err
		}

		newRefreshToken, err := randomToken(32)
		if err != nil {
			return err
		}
		rotatedToken, err := s.authRepository.RotateRefreshToken(
			txCtx,
			currentToken.ID,
			currentTokenHash,
			hashRefreshToken(newRefreshToken),
			time.Now().UTC().Add(s.cfg.RefreshTokenTTL),
		)
		if err != nil {
			return err
		}

		accessToken, err := s.issueAccessToken(rotatedToken.UserID, rotatedToken.ActiveRole, rotatedToken.ID)
		if err != nil {
			return err
		}
		tokens = &entity.JWT{AccessToken: accessToken, RefreshToken: newRefreshToken}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	if strings.TrimSpace(refreshToken) == "" {
		return entity.ErrRefreshTokenMissing
	}
	return s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		currentToken, err := s.authRepository.GetRefreshTokenByHashForUpdate(txCtx, hashRefreshToken(refreshToken))
		if err != nil {
			return err
		}
		if err := validateRefreshTokenState(currentToken); err != nil {
			return err
		}
		return s.authRepository.RevokeRefreshToken(txCtx, currentToken.ID)
	})
}

func (s *authService) IssueAccessToken(_ context.Context, userID string, role entity.ProductRole, sessionID string) (string, error) {
	return s.issueAccessToken(userID, role, sessionID)
}

func (s *authService) VerifyAccessToken(_ context.Context, token string) (*entity.AccessClaims, error) {
	if strings.TrimSpace(token) == "" {
		return nil, entity.ErrAccessTokenMissing
	}
	return s.parseAccessToken(token)
}

func (s *authService) createTokenPair(ctx context.Context, userID string, role entity.ProductRole) (*entity.JWT, error) {
	rawRefreshToken, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	sessionID, err := entity.NewUUID()
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.authRepository.CreateRefreshToken(ctx, entity.RefreshToken{
		ID:         sessionID,
		UserID:     userID,
		TokenHash:  hashRefreshToken(rawRefreshToken),
		ActiveRole: role,
		ExpiresAt:  time.Now().UTC().Add(s.cfg.RefreshTokenTTL),
	})
	if err != nil {
		return nil, err
	}
	accessToken, err := s.issueAccessToken(userID, role, refreshToken.ID)
	if err != nil {
		return nil, err
	}
	return &entity.JWT{AccessToken: accessToken, RefreshToken: rawRefreshToken}, nil
}

func (s *authService) issueAccessToken(userID string, role entity.ProductRole, sessionID string) (string, error) {
	now := time.Now().UTC()
	claims := accessTokenClaims{
		UserID:     userID,
		ActiveRole: string(role),
		SessionID:  sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.cfg.JWTSecret)
}

func (s *authService) parseAccessToken(token string) (*entity.AccessClaims, error) {
	claims := &accessTokenClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return s.cfg.JWTSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, entity.ErrAccessTokenExpired
		}
		return nil, entity.ErrAccessTokenInvalid
	}
	if parsed == nil || !parsed.Valid {
		return nil, entity.ErrAccessTokenInvalid
	}
	role := entity.ProductRole(claims.ActiveRole)
	if claims.UserID == "" || claims.SessionID == "" || !role.Valid() || claims.Subject != claims.UserID {
		return nil, entity.ErrAccessTokenInvalid
	}
	return &entity.AccessClaims{
		UserID:     claims.UserID,
		ActiveRole: role,
		SessionID:  claims.SessionID,
		ExpiresAt:  claims.ExpiresAt.Time,
	}, nil
}

func validateRefreshTokenState(token *entity.RefreshToken) error {
	if token == nil {
		return entity.ErrRefreshTokenInvalid
	}
	if !time.Now().UTC().Before(token.ExpiresAt) {
		return entity.ErrRefreshTokenExpired
	}
	if token.RevokedAt != nil {
		return entity.ErrRefreshTokenRevoked
	}
	return nil
}

func roleInList(role entity.ProductRole, roles []entity.ProductRole) bool {
	for _, item := range roles {
		if item == role {
			return true
		}
	}
	return false
}

func randomToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifyPassword(encoded, password string) bool {
	if strings.HasPrefix(encoded, "$2a$") || strings.HasPrefix(encoded, "$2b$") || strings.HasPrefix(encoded, "$2y$") {
		return bcrypt.CompareHashAndPassword([]byte(encoded), []byte(password)) == nil
	}
	return false
}
