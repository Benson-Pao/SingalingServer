package jwt

import (
	"log"

	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

var secretKey string = "d97ff91633ae386905cfe4230361fb0eb7c1083aba29ddf1ce21bbad3339fa82"

// 模擬生成有效的 JWT
func generateValidJWT(secretKey string) string {
	token, _ := GenerateJWT(1, "1", "1", []byte(secretKey))
	log.Printf("生成的 JWT: %s", token)
	return token
}

// 解析 JWT
func ValidateJWT(tokenString string, secretKey []byte) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	// 驗證並提取 claims
	if claims, ok := token.Claims.(*jwt.MapClaims); ok && token.Valid {
		return *claims, nil
	}
	return nil, jwt.NewValidationError("invalid token", jwt.ValidationErrorClaimsInvalid)
}
func TestJWTValidation(t *testing.T) {
	// 生成有效的 JWT
	token := generateValidJWT(secretKey)

	// 解析 JWT
	claims, err := ParseClaims(token, []byte(secretKey))
	assert.NoError(t, err, "JWT 應該驗證成功")

	// 驗證用戶 ID，這裡進行類型轉換
	userID, ok := claims["sub"].(float64) // 用戶 ID 在 JWT 中通常會解析為 float64
	assert.True(t, ok, "用戶 ID 應該為 float64 類型")
	assert.Equal(t, float64(1), userID, "用戶 ID 應該為 1") // 預期值是 float64 類型
}
