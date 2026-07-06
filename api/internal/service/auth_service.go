package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
)

// ErrInvalidCredentials est renvoyée quand le couple username/password est incorrect, sans préciser lequel, pour ne pas aider une attaque par énumération.
var ErrInvalidCredentials = errors.New("identifiants invalides")

// Claims représente le contenu embarqué dans le JWT émis à la connexion.
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthService porte la logique d'inscription, de connexion et de vérification des tokens JWT.
type AuthService struct {
	users      *repository.UserRepository
	jwtSecret  []byte
	expiration time.Duration
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, expiration time.Duration) *AuthService {
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret), expiration: expiration}
}

// Register crée un nouvel utilisateur avec un mot de passe hashé (bcrypt). Le rôle par défaut est "employee" ; il n'y a pas de création d'admin via cette route publique, conformément à un usage boutique mono-équipe.
func (s *AuthService) Register(username, password string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hachage du mot de passe : %w", err)
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "employee",
	}
	id, err := s.users.Create(user)
	if err != nil {
		return nil, err
	}

	return s.users.FindByID(id)
}

// Login vérifie les identifiants et génère un JWT signé valable pour la durée configurée (JWT_EXPIRATION_HOURS).
func (s *AuthService) Login(username, password string) (token string, expiresAt time.Time, user *models.User, err error) {
	user, err = s.users.FindByUsername(username)
	if errors.Is(err, repository.ErrNotFound) {
		return "", time.Time{}, nil, ErrInvalidCredentials
	}
	if err != nil {
		return "", time.Time{}, nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", time.Time{}, nil, ErrInvalidCredentials
	}

	expiresAt = time.Now().Add(s.expiration)
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, nil, fmt.Errorf("génération du token : %w", err)
	}

	return token, expiresAt, user, nil
}

// ValidateToken vérifie la signature et l'expiration d'un JWT et retourne ses claims. Utilisée par le middleware d'authentification.
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("méthode de signature inattendue : %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("token invalide : %w", err)
	}
	return claims, nil
}
