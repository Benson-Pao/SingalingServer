
# Signaling Server with JWT Authentication and WebRTC Communication

## 簡介
本專案主要功能為信令伺服器，以WebRTC 將畫面擷取為影音傳送至另一個使用者做為影音傳遞，利用WebSocket做為信令傳送達成WebRTC信令交換使用，並透過JWT為web與WebSocket驗證存取權限。
```
此段程試因測試方便用僅單一影片提供者(為第一個戴入sender的user)
若要變動可修改下列
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
```
```
若想自訂影片源提供者改成下列

    //msg.UserMessage.Target (在js指定接收者id)
	msg.UserMessage.Sender = clientInfo.UID

	//因原始傳來的結構沒type所以附加
	msg.Offer.Type = msg.Type
	castMessage(clientInfo, msg)

            
```

此專案的主要功能是：

- 提供 WebRTC 訊號交換流程（`offer` 和 `answer`）來建立視頻通話。
- 使用 WebSocket 作為訊號伺服器來處理客戶端之間的連線。
- 利用 JWT 進行身份驗證，並提供 token 驗證和刷新功能。

專案因測試用有兩個主要頁面：
1. `receiver.htm`：視頻接收端，發送 `offer` 請求並接收 `answer` 回應。
2. `sender.htm`：視頻發送端，接收 `offer` 並回應 `answer`。

### WebRTC 流程

1. **receiver.htm** 發送 `offer` 訊號給 **sender.htm**（視頻發送端）。
2. **sender.htm** 接收到 `offer` 訊號後，發送 `answer` 訊號回應。
3. 雙方成功建立 WebRTC 連線並開始視頻通話。

## 使用說明

### 啟動伺服器

1. 安裝專案所需的依賴包：

   確保安裝了 `gin` 和 `ws` 等依賴包。

2. 使用以下指令啟動伺服器：

   go run main.go -k YOUR_SECRET_KEY -p PORT -d DOMAIN



這個專案提供了一個簡單的信令伺服器，實現了 WebRTC 通訊與 JWT 驗證。該伺服器協助兩個用戶端（發送端和接收端）建立 WebRTC 連線，並使用 JWT 來處理身份驗證，並實作擷取畫面傳送給另一接接收的簡易測試。

## 簡介

此專案的主要功能是：

- 提供 WebRTC 訊號交換流程（`offer` 和 `answer`）來建立視頻通話。
- 使用 WebSocket 作為訊號伺服器來處理客戶端之間的連線。
- 利用 JWT 進行身份驗證，並提供 token 驗證和刷新功能。

專案中有兩個主要頁面：
1. `receiver.htm`：視頻接收端，發送 `offer` 請求並接收 `answer` 回應。
2. `sender.htm`：視頻發送端，接收 `offer` 並回應 `answer`。

### WebRTC 流程

1. **receiver.htm** 發送 `offer` 訊號給 **sender.htm**（視頻發送端）。
2. **sender.htm** 接收到 `offer` 訊號後，發送 `answer` 訊號回應。
3. 雙方成功建立 WebRTC 連線並開始視頻通話。

## 使用說明

### 啟動伺服器
1. 使用以下指令啟動伺服器：
   go run main.go -k YOUR_SECRET_KEY -p PORT -d DOMAIN

-   `-k YOUR_SECRET_KEY`：設定 JWT 的密鑰，若未提供將使用預設密鑰。
-   `-p PORT`：設定伺服器的端口，默認為 `:8080`。
-   `-d DOMAIN：設定伺服器的域骀，默認為 `localhost。

2.  訪問以下頁面來進行視頻通話：
	 -   Sender  ex http://localhost:8080/sender (影片提供者 需先啟動)
    -   Receiver ex http://localhost:8080/receiver (影片接收者)

  

### API 端點

-   **`/ws`**：WebSocket 連接，用於交換 WebRTC 訊號（`offer`、`answer`、`candidate`）。
    
-   **`/create/:role`**：生成 Access Token 和 Refresh Token，並根據角色進行身份驗證。
    
    -   `role`：可為 `1為sender` 或 `0為receiver`，用於區分視頻發送端和接收端。
    -   返回的 JSON 包含：
        -   `access_token`：有效的 Access Token，用於 API 請求。
        -   `refresh_token`：用於刷新 Access Token 的 Refresh Token。
-   **`/first`**：檢查是否為第一次發送者，回傳 `true` 或 `false`。
    
-   **`/refresh`**：使用 Refresh Token 生成新的 Access Token。
    

### JWT 參數說明

-   **Access Token**：用於身份驗證的令牌。有效期較短，需定期刷新。
-   **Refresh Token**：用於刷新 Access Token 的令牌，當 Access Token 過期時，使用 Refresh Token 獲得新的 Access Token。

### WebRTC 訊號流程

1.  **receiver.htm** 發送 `offer` 訊號給 **sender.htm**，開始連線請求。
2.  **sender.htm** 接收到 `offer`，發送 `answer` 訊號回應。
3.  雙方透過 WebRTC 開始視頻通話。

## 參數說明

### 啟動參數

-   -k YOUR_SECRET_KEY`：設定 JWT 密鑰，用於生成和驗證令牌。
-   -p PORT`：設定伺服器端口，默認為 `:8080 (:要打)
-   -d  domain : 設定域名，默認為localhost


### WebSocket 訊號

-   `offer`：視頻接收端發送的連線請求，包含 WebRTC 的媒體協商資料。
-   `answer`：視頻發送端回應的資料，包含接收端協商過的媒體資訊。
-   `candidate`：WebRTC 中的 ICE candidate，表示可能的網路路徑。

## 目錄結構`.
```

├── main.go                 # Go 設定 WebSocket 與 HTTP 伺服器
├── view/
│   ├── receiver.htm        # 接收端測試頁面 (視頻接收者)
│   ├── sender.htm          # 發送端測試頁面 (視頻發送者)
├── js/
│   ├── webrtc.js           # WebRTC 邏輯 (建立連線與訊號處理)
│   └── lock.js             # js lock
├──-conn/                
│		└── ws.go			# 信令相關處理 (WebSocket)
├── util/                           
│	└── Sonyflake.go	    # 生成 ID
├── jwt/
	└── token.go		    # JWT

```
## 注意事項

-   **WebSocket 連接**：請確保伺服器已啟動並正確配置，才能讓客戶端進行信號交換。
-   **JWT 密鑰**：務必妥善保管密鑰，確保身份驗證過程的安全。
-   **WebRTC 訊號交換**：`offer` 和 `answer` 訊號交換必須正確進行，否則無法建立有效的視頻通話。

## 授權

此專案使用 MIT 授權協議。詳情請參閱 LICENSE 檔案。

