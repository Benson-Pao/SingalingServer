package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func GenerateSecretKey() string {
	// 生成 256 位隨機密鑰
	secretKey := make([]byte, 32) // 256 位
	_, err := rand.Read(secretKey)
	if err != nil {
		fmt.Println("Error generating random key:", err)
		return ""
	}

	return hex.EncodeToString(secretKey)
}

// 用於驗證 token 是否有效
func IsValidToken(tokenStr string, secretKey []byte) bool {
	// 檢查 token 是否為空
	if tokenStr == "" {
		return false
	}

	// 解析 token
	token, err := ParseToken(tokenStr, secretKey)
	if err != nil {
		log.Println("解析 Token 失敗:", err)
		return false
	}

	// 驗證 token 是否有效
	// 確保 token 包含有效的聲明並且 token 本身是有效的
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Println("無效的 Token")
		return false
	}

	// 檢查過期時間 (exp) 是否過期
	expirationTime := claims["exp"].(float64) // exp 是 Unix 時間戳
	if expirationTime < float64(time.Now().Unix()) {
		log.Println("Token 已過期")
		return false
	}

	return true
}

// 用來解析 JWT token 並返回 token
func ParseToken(tokenStr string, secretKey []byte) (*jwt.Token, error) {
	// 解析 token 並驗證
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// 確保 token 使用正確的簽名算法
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("不支援的簽名方法 %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	return token, err
}

// 用來生成 JWT token
func GenerateJWT(userID uint64, userName string, role string, secretKey []byte) (string, error) {

	// 從環境變數中讀取過期時間，若未設定則默認為 1 小時
	expirationHours := os.Getenv("JWT_EXPIRATION_HOURS")
	if expirationHours == "" {
		expirationHours = "1" // 預設 1 小時
	}

	hours, err := strconv.ParseInt(expirationHours, 10, 64)
	if err != nil {
		log.Println("無法解析過期時間，使用默認值 1 小時")
		hours = 1
	}
	expirationTime := time.Now().Add(time.Hour * time.Duration(hours)).Unix()

	claims := jwt.MapClaims{
		"sub":  userID,            // 用戶 ID
		"name": userName,          // 用戶名稱
		"iat":  time.Now().Unix(), // 發行時間
		"exp":  expirationTime,    // 過期
		"role": role,              // 0接收影片者 1發佈影片者
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("生成 Token 失敗: %v", err)
	}

	return tokenString, nil
}

// 用來生成 JWT refresh token
func GenerateRefreshJWT(userID uint64, userName string, role string, secretKey []byte) (string, error) {
	// 從環境變數中讀取過期時間，若未設定則默認為 30 天
	expirationDaysStr := os.Getenv("JWT_REFRESH_EXPIRATION_DAYS")
	if expirationDaysStr == "" {
		expirationDaysStr = "30" // 預設 30 天
	}

	// 解析過期時間為 int64
	days, err := strconv.ParseInt(expirationDaysStr, 10, 64)
	if err != nil {
		log.Println("無法解析過期時間，使用默認值 30 天")
		days = 30
	}

	// 計算過期時間
	expirationTime := time.Now().Add(time.Duration(days) * 24 * time.Hour).Unix()

	claims := jwt.MapClaims{
		"sub":     userID,            // 用戶 ID
		"name":    userName,          // 用戶名稱
		"iat":     time.Now().Unix(), // 發行時間
		"exp":     expirationTime,    // 過期時間
		"role":    role,              // 0接收影片者 1發佈影片者
		"refresh": true,              // 標記為 refresh token
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("生成 Refresh Token 失敗: %v", err)
	}

	return tokenString, nil
}

// 用來解析 JWT Token 並返回聲明
func ParseClaims(tokenStr string, secretKey []byte) (jwt.MapClaims, error) {
	// 解析 token
	token, err := ParseToken(tokenStr, secretKey)
	if err != nil {
		return nil, err
	}

	// 返回解析後的聲明
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("無效的 Token")
	}

	return claims, nil
}

// 這隻取得Token (ws/http) 本身不騳證 依IsValidToken 來驗證
func GetToken(r *http.Request) string {

	token := r.Header.Get("Authorization")

	if token == "" {
		token = r.URL.Query().Get("token")
		return token
	}

	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}
	return token
}
