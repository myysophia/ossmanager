package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/myysophia/ossmanager-backend/internal/config"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"time"
)

// 定义错误
var (
	ErrInvalidToken = errors.New("令牌无效")
	ErrExpiredToken = errors.New("令牌已过期")
)

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(user *models.User, jwtConfig *config.JWTConfig) (string, error) {
	// 设置过期时间
	expiresAt := time.Now().Add(jwtConfig.GetJWTExpiration())

	// 创建声明
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    jwtConfig.Issuer,
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(jwtConfig.SecretKey))
	if err != nil {
		return "", fmt.Errorf("生成令牌失败: %w", err)
	}

	return tokenString, nil
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string, jwtConfig *config.JWTConfig) (*Claims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtConfig.SecretKey), nil
	})

	if err != nil {
		// 检查是否是过期错误
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrExpiredToken
			}
		}
		return nil, ErrInvalidToken
	}

	// 获取声明
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
