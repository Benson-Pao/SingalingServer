package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	ws "singaling/conn"
	"singaling/jwt"
	auth "singaling/jwt"
	"singaling/util"

	"github.com/gin-gonic/gin"
)

var (
	secretKey string
	port      string
	domain    string
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	flag.StringVar(&secretKey, "k", "d97ff91633ae386905cfe4230361fb0eb7c1083aba29ddf1ce21bbad3339fa82", "JWT 簽名的密鑰")
	flag.StringVar(&port, "p", ":8080", "Web監聽Port")
	flag.StringVar(&domain, "d", "localhost", "Web Domain")
	flag.Parse()

	// 使用環境變數讀取密鑰(環境變數優先)
	if os.Getenv("SECRET_KEY") != "" {
		secretKey = os.Getenv("SECRET_KEY")
	}

	// 防止用戶設置空值
	if secretKey == "" {
		log.Fatal("密鑰不能為空，請通過 -key 提供有效的密鑰")
	}

	// 防止用戶設置domain空值
	if domain == "" {
		log.Fatal("domain不能為空，請通過 -d 提供有效的domain")
	}

	// 防止用戶設置Port空值
	if port == "" {
		log.Fatal("Port不能為空，請通過 -p 提供有效的Port")
	}
}

func main() {

	fmt.Println("當前使用的密鑰為:", secretKey)

	// Gin 路由
	r := gin.Default()

	// 呼叫 auth 包中的 GenerateSecretKey 函數來生成隨機密鑰
	//key := auth.GenerateSecretKey()
	//fmt.Println("Generated Key:", key)

	// 設置靜態文件路由
	r.Static("/js", "./js")

	// 設置模板路由
	r.LoadHTMLFiles("view/sender.htm",
		"view/receiver.htm")

	fmt.Printf("WebRTC Receiver  http://localhost%s/receiver\n", port)
	fmt.Printf("WebRTC Sender  http://localhost%s/sender\n", port)

	r.GET("/sender", func(c *gin.Context) {
		// sender.htm 頁面
		c.HTML(http.StatusOK, "sender.htm", gin.H{
			"domain": domain,
			"port":   port,
		})
	})

	r.GET("/receiver", func(c *gin.Context) {
		// receiver.htm 頁面
		c.HTML(http.StatusOK, "receiver.htm", gin.H{
			"domain": domain,
			"port":   port,
		})
	})

	// 啟動廣播 Goroutine
	// WebSocket 路由，調用 ws 包的 HandleWebSocket 函數
	r.GET("/ws", ws.HandleWebSocket(secretKey))

	r.GET("/create/:role", func(c *gin.Context) {
		role := c.Param("role")
		log.Println(role)
		if role == "" {
			role = "0"
		}
		//照理說要登入流程取得id，此處簡化
		userID, err := util.GenerateID() // 用雪花 ID 作為 UserID
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 UserID 失敗"})
			return
		}

		// 使用雪花 ID 和名字生成 Access Token 和 Refresh Token
		accessToken, err := auth.GenerateJWT(userID, strconv.FormatUint(userID, 10), role, []byte(secretKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Access Token 失敗"})
			return
		}

		refreshToken, err := auth.GenerateRefreshJWT(userID, strconv.FormatUint(userID, 10), role, []byte(secretKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Refresh Token 失敗"})
			return
		}

		// 返回生成的 Access Token 和 Refresh Token
		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})

	r.GET("/first", func(c *gin.Context) {
		token := jwt.GetToken(c.Request)
		log.Println("Access Token:", token)

		if !jwt.IsValidToken(token, []byte(secretKey)) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		info := ws.GetFirstSenderInfo()
		if info == nil {
			log.Println("First sender info not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "First sender not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"first": info.UID})
	})

	// 路由處理用於生成新的 Access Token
	r.POST("/refresh", func(c *gin.Context) {
		// 從請求的 body 中設定 RefreshToken
		var requestData struct {
			RefreshToken string `json:"refresh_token"` // 提取 Access Token
		}

		// 解析請求的 body
		if err := c.ShouldBindJSON(&requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無效的請求格式"})
			return
		}

		// 檢查 refresh_token 是否有效
		valid := auth.IsValidToken(requestData.RefreshToken, []byte(secretKey))
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh Token 無效"})
			return
		}

		// 從 Refresh Token 中解析出用戶 ID 和名稱
		claims, err := auth.ParseClaims(requestData.RefreshToken, []byte(secretKey))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "無法解析 Refresh Token"})
			return
		}

		userID := uint64(claims["sub"].(float64)) // 解析出用戶 ID
		userName := claims["name"].(string)       // 解析出用戶名稱
		role := claims["role"].(string)           // 解析出role

		// 使用用戶 ID 和名稱生成新的 Access Token
		newAccessToken, err := auth.GenerateJWT(userID, userName, role, []byte(secretKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成新的 Access Token 失敗"})
			return
		}

		// 返回新的 Access Token
		c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
	})

	fmt.Printf("信令伺服器運行中，連接 ws://localhost%s/ws\n", port)
	if err := r.Run(port); err != nil {
		log.Fatal("伺服器啟動失敗:", err)
	}
}
