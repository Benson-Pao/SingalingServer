package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"singaling/jwt"

	"singaling/util"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	// 使用 sync.Map 來避免並發讀寫競爭
	clients sync.Map
)

type Message struct {
	Type        string       `json:"type"` // offer, answer, candidate
	Data        interface{}  `json:"data"`
	Token       string       `json:"token"`
	Offer       OfferMsg     `json:"offer,omitempty"`
	Candidate   CandidateMsg `json:"candidate,omitempty"`
	UserMessage UserMessage  `json:"userMessage"`
}

type OfferMsg struct {
	Type string `json:"type"` // offer, answer, candidate
	SDP  string `json:"sdp"`
}

type CandidateMsg struct {
	Candidate     string `json:"candidate"`
	SdpMid        string `json:"sdpMid"`
	SdpMLineIndex int    `json:"sdpMLineIndex"`
}

type UserMessage struct {
	Target string `json:"target"` // 目標使用者
	Sender string `json:"sender"` // 發送者
}

type ClientInfo struct {
	UID           string
	ConnID        uint64
	Conn          *websocket.Conn
	LastActive    time.Time
	HeartbeatChan chan bool
	TargetID      string
	Role          string //0訂閱影音者 1發送影音者
	Ctx           context.Context
	IsConnecting  bool
	Lock          sync.Mutex
}

func (c *ClientInfo) Close() {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	if c.IsConnecting {
		c.Conn.Close()
		//close(c.HeartbeatChan)
		c.IsConnecting = false
	}
}

// WebSocket 連線處理
func HandleWebSocket(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {

		upgrader := &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				token := jwt.GetToken(r)
				log.Println("Access Token:", token)
				return jwt.IsValidToken(token, []byte(secretKey))
			},
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("WebSocket 升級失敗:", err)
			return
		}
		defer ws.Close()

		//基本上過了前面這裡就一定有資料 若驗證不過Upgrade會擋下來
		claims, _ := jwt.ParseClaims(jwt.GetToken(c.Request), []byte(secretKey))
		userID, _ := claims["sub"].(float64)
		uid := strconv.FormatFloat(userID, 'f', -1, 64)
		connID, _ := util.GenerateID()
		role, _ := claims["role"].(string)

		// 創建一個單獨的 context 用於這個連線的生命週期
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		clientInfo := &ClientInfo{
			UID:           uid,
			ConnID:        connID,
			Conn:          ws,
			LastActive:    time.Now(),
			HeartbeatChan: make(chan bool, 10),
			TargetID:      "",
			Ctx:           ctx,
			Role:          role,
			IsConnecting:  true,
		}

		old, ok := clients.LoadAndDelete(clientInfo.UID)
		if ok {
			oldclient := old.(*ClientInfo)
			oldclient.Close()
		}

		// 將客戶端加入 clients，並確保操作是原子性的
		clients.Store(clientInfo.UID, clientInfo)
		log.Printf("新客戶端連線 UserID:%s ConnID:%d\n", uid, connID)

		// 開始監控該客戶端的心跳
		go monitorHeartbeat(clientInfo)
		defer func() {
			close(clientInfo.HeartbeatChan)
		}()

		// 處理訊息
		for {

			messageType, p, err := ws.ReadMessage()
			if err != nil {
				log.Printf("錯誤，%s關閉連線: %+v\n", clientInfo.UID, err)
				// 客戶端失敗時從 clients 中刪除
				clients.Delete(clientInfo.UID)
				cancel()
				break
			}

			// 確保訊息類型正確（文本類型）
			if messageType != websocket.TextMessage {
				log.Printf("非文本類型訊息，忽略: %d", messageType)
				continue
			}

			var msg Message
			//err := ws.ReadJSON(&msg)
			err = json.Unmarshal(p, &msg)
			if err != nil {
				//log.Printf("JSON 解析失敗，原始數據: %s，錯誤: %v", string(p), err)
				if syntaxErr, ok := err.(*json.SyntaxError); ok {
					log.Printf("JSON 語法錯誤: Offset %d, 原始數據: %s", syntaxErr.Offset, string(p))
				} else if typeErr, ok := err.(*json.UnmarshalTypeError); ok {
					log.Printf("JSON 類型錯誤: Value %s, Type %s, 原始數據: %s", typeErr.Value, typeErr.Type, string(p))
				} else {
					log.Printf("JSON 解析失敗，原始數據: %s，錯誤: %v", string(p), err)
				}
				continue
			}

			//jsonbytes, _ := json.Marshal(msg)
			log.Printf("刷新活躍時間及送出廣播:UID %s Message:%+v\n", uid, string(p))
			clientInfo.HeartbeatChan <- true

			//檢查候選訊息是否包含必要的屬性
			switch msg.Type {
			case "candidate":

				//因原始傳來的結構沒type所以附加
				msg.Offer.Type = msg.Type
				msg.UserMessage.Sender = clientInfo.UID
				castMessage(clientInfo, msg)

			case "offer":
				//因測試只一個接收者
				receiver := GetFirstSenderInfo()
				if receiver != nil {
					msg.UserMessage.Target = receiver.UID
					msg.UserMessage.Sender = clientInfo.UID

					//因原始傳來的結構沒type所以附加
					msg.Offer.Type = msg.Type
					castMessage(clientInfo, msg)
				}
			case "answer":

				msg.Offer.Type = msg.Type
				msg.UserMessage.Sender = clientInfo.UID
				castMessage(clientInfo, msg)

			}

		}
	}
}

func castMessage(curUser *ClientInfo, msg Message) {
	log.Println(msg.UserMessage.Target)
	if client, ok := clients.Load(msg.UserMessage.Target); ok {
		if value, ok := client.(*ClientInfo); ok {
			if err := writeMessage(value.Conn, msg); err != nil {
				log.Printf("%s無法發送訊息至%s err%s\n", curUser.UID, msg.UserMessage.Target, err)
			}
		}
	} else {
		log.Printf("%s無法發送訊息至%s 用戶 %s 不在線\n", curUser.UID, msg.UserMessage.Target, msg.UserMessage.Target)
	}

}

func GetSender() []string {
	var senders []string
	clients.Range(func(key, value interface{}) bool {
		clientInfo, ok := value.(*ClientInfo)
		if ok && clientInfo.Role == "1" {
			senders = append(senders, clientInfo.UID)
		}
		return true // 繼續遍歷
	})
	return senders
}

func GetFirstSenderInfo() *ClientInfo {
	var c *ClientInfo
	clients.Range(func(key, value interface{}) bool {
		clientInfo, ok := value.(*ClientInfo)
		log.Printf("client %+v\n", clientInfo)
		if ok && clientInfo.Role == "1" {
			c = clientInfo
			return false
		}
		return true
	})
	return c
}

// 廣播訊息給所有客戶端
func broadcastMessage(curUser *ClientInfo, msg any) {
	// 使用 sync.Map 的 Range 來遍歷所有客戶端
	clients.Range(func(key, value interface{}) bool {
		// 直接從 value 取得 *ClientInfo
		clientInfo := value.(*ClientInfo)
		// 從 ClientInfo 取得 websocket.Conn
		client := clientInfo.Conn

		if curUser.Conn != client {
			// 發送訊息並處理錯誤
			err := writeMessage(client, msg)
			if err != nil {
				log.Printf("訊息發送失敗，客戶端%s 已斷線:%s\n", clientInfo.UID, err)
				// 當訊息發送失敗時，關閉連線並刪除該客戶端
				client.Close()
				clients.Delete(client)
			} else {
				// 成功發送訊息後更新該客戶端的活動時間
				clientInfo.LastActive = time.Now()
			}
		}
		return true
	})
}

// 寫入訊息到 WebSocket 連線的函數
func writeMessage(client *websocket.Conn, msg any) error {
	//因此套件的寫入內鍵己有加鎖所以避免競爭即不用另行加鎖了(可參見原和程式)
	err := client.WriteJSON(msg)
	return err
}

// 監控心跳
func monitorHeartbeat(clientInfo *ClientInfo) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// 設定最大閒置時間為 5 分鐘
	maxIdleDuration := 5 * time.Minute
	lastActive := time.Now()

	for {
		select {
		case <-ticker.C:
			if err := clientInfo.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("連線閒置過久，客戶端%s Ping 發送失敗\n", clientInfo.UID)
				clients.Delete(clientInfo.UID)
				clientInfo.Close()
				return
			}
			//log.Printf("使用者%+v Ping 發送完成", clientInfo.UID)

		case <-clientInfo.HeartbeatChan:
			// 更新最後活躍時間
			clientInfo.LastActive = time.Now()
			lastActive = clientInfo.LastActive

		case <-time.After(maxIdleDuration):
			// 檢查閒置時間，如果超過最大閒置時間，關閉連線
			if time.Since(lastActive) >= maxIdleDuration {
				log.Printf("連線閒置過久，客戶端%s 關閉連線\n", clientInfo.UID)
				clients.Delete(clientInfo.UID)
				clientInfo.Close()
				return
			}
		case <-clientInfo.Ctx.Done():
			return
		}
	}
}
