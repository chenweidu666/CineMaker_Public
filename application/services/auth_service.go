package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db  *gorm.DB
	cfg *config.Config
	log *logger.Logger
}

func NewAuthService(db *gorm.DB, cfg *config.Config, log *logger.Logger) *AuthService {
	return &AuthService{db: db, cfg: cfg, log: log}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=128"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	User         models.User `json:"user"`
}

type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	TeamID    uint   `json:"team_id"`
	Role      string `json:"role"`
	TokenType string `json:"typ"`
	jwt.RegisteredClaims
}

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

func (s *AuthService) Register(req *RegisterRequest) (*TokenResponse, error) {
	var existing models.User
	if err := s.db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, errors.New("该邮箱已被注册")
	}
	if err := s.db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, errors.New("该用户名已被使用")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost+2)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	var userID uint
	var teamID uint

	err = s.db.Transaction(func(tx *gorm.DB) error {
		user := &models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: string(hash),
			Role:         "owner",
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		userID = user.ID

		team := &models.Team{
			Name:    req.Username + " 的团队",
			OwnerID: user.ID,
		}
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		teamID = team.ID

		return tx.Model(user).Update("team_id", team.ID).Error
	})
	if err != nil {
		return nil, fmt.Errorf("注册失败: %w", err)
	}

	var user models.User
	if err := s.db.Preload("Team").First(&user, userID).Error; err != nil {
		s.log.Errorw("Failed to load user after registration", "error", err, "user_id", userID, "team_id", teamID)
		return nil, fmt.Errorf("注册成功但加载用户信息失败")
	}

	return s.generateTokenResponse(&user)
}

func (s *AuthService) Login(req *LoginRequest) (*TokenResponse, error) {
	var user models.User
	if err := s.db.Preload("Team").Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, errors.New("邮箱或密码错误")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("邮箱或密码错误")
	}

	return s.generateTokenResponse(&user)
}

func (s *AuthService) RefreshToken(refreshToken string) (*TokenResponse, error) {
	claims, err := s.ParseToken(refreshToken, TokenTypeRefresh)
	if err != nil {
		return nil, errors.New("无效的刷新令牌")
	}

	var user models.User
	if err := s.db.Preload("Team").First(&user, claims.UserID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	return s.generateTokenResponse(&user)
}

func (s *AuthService) GetCurrentUser(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Team").First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func (s *AuthService) UpdateUser(userID uint, username, avatar string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	updates := map[string]interface{}{}
	if username != "" {
		var existing models.User
		if err := s.db.Where("username = ? AND id != ?", username, userID).First(&existing).Error; err == nil {
			return nil, errors.New("该用户名已被使用")
		}
		updates["username"] = username
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}

	if len(updates) > 0 {
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, err
		}
	}

	s.db.Preload("Team").First(&user, userID)
	return &user, nil
}

func (s *AuthService) ParseToken(tokenStr string, expectedType string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWT.Secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if expectedType != "" && claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s", expectedType)
	}
	return claims, nil
}

// Team operations

func (s *AuthService) GetTeam(teamID uint) (*models.Team, error) {
	var team models.Team
	if err := s.db.Preload("Members").Preload("Owner").First(&team, teamID).Error; err != nil {
		return nil, errors.New("团队不存在")
	}
	return &team, nil
}

func (s *AuthService) UpdateTeam(teamID, userID uint, name string) (*models.Team, error) {
	var team models.Team
	if err := s.db.First(&team, teamID).Error; err != nil {
		return nil, errors.New("团队不存在")
	}
	if team.OwnerID != userID {
		return nil, errors.New("只有团队所有者可以修改团队信息")
	}
	team.Name = name
	if err := s.db.Save(&team).Error; err != nil {
		return nil, err
	}
	s.db.Preload("Members").Preload("Owner").First(&team, teamID)
	return &team, nil
}

func (s *AuthService) InviteMember(teamID, userID uint, email string) (*models.Invitation, error) {
	var team models.Team
	if err := s.db.First(&team, teamID).Error; err != nil {
		return nil, errors.New("团队不存在")
	}
	if team.OwnerID != userID {
		return nil, errors.New("只有团队所有者可以邀请成员")
	}

	var existingUser models.User
	if err := s.db.Where("email = ? AND team_id = ?", email, teamID).First(&existingUser).Error; err == nil {
		return nil, errors.New("该用户已是团队成员")
	}

	inv := &models.Invitation{
		TeamID:    teamID,
		Email:     email,
		Token:     uuid.New().String(),
		Status:    "pending",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := s.db.Create(inv).Error; err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *AuthService) AcceptInvitation(token string, userID uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		var inv models.Invitation
		if err := tx.Where("token = ? AND status = ?", token, "pending").First(&inv).Error; err != nil {
			return errors.New("邀请不存在或已过期")
		}
		if time.Now().After(inv.ExpiresAt) {
			tx.Model(&inv).Update("status", "expired")
			return errors.New("邀请已过期")
		}
		if user.Email != inv.Email {
			return errors.New("该邀请不属于当前用户")
		}

		if err := tx.Model(&models.User{}).Where("id = ?", userID).
			Updates(map[string]interface{}{"team_id": inv.TeamID, "role": "member"}).Error; err != nil {
			return err
		}
		return tx.Model(&inv).Update("status", "accepted").Error
	})
}

func (s *AuthService) RemoveMember(teamID, ownerID, memberID uint) error {
	var team models.Team
	if err := s.db.First(&team, teamID).Error; err != nil {
		return errors.New("团队不存在")
	}
	if team.OwnerID != ownerID {
		return errors.New("只有团队所有者可以移除成员")
	}
	if memberID == ownerID {
		return errors.New("不能移除自己")
	}
	result := s.db.Model(&models.User{}).Where("id = ? AND team_id = ?", memberID, teamID).
		Updates(map[string]interface{}{"team_id": nil, "role": "member"})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("该成员不在团队中")
	}
	return nil
}

func (s *AuthService) generateTokenResponse(user *models.User) (*TokenResponse, error) {
	var teamID uint
	if user.TeamID != nil {
		teamID = *user.TeamID
	}

	accessTTL := time.Duration(s.cfg.JWT.AccessTokenTTL) * time.Minute
	if accessTTL == 0 {
		accessTTL = 30 * time.Minute
	}
	refreshTTL := time.Duration(s.cfg.JWT.RefreshTokenTTL) * 24 * time.Hour
	if refreshTTL == 0 {
		refreshTTL = 7 * 24 * time.Hour
	}

	accessToken, err := s.generateJWT(user.ID, teamID, user.Role, TokenTypeAccess, accessTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.generateJWT(user.ID, teamID, user.Role, TokenTypeRefresh, refreshTTL)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTTL.Seconds()),
		User:         *user,
	}, nil
}

func (s *AuthService) generateJWT(userID, teamID uint, role string, tokenType string, ttl time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID:    userID,
		TeamID:    teamID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}
