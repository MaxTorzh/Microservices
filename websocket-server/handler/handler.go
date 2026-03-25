package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"websocket_server/client"
	"websocket_server/server"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Обработка WebSocket подключения
type WebSocketHandler struct {
	server   *server.Server
	upgrader websocket.Upgrader
}

// Создание нового WebSocket обработчика
func NewWebSocketHandler(s *server.Server) *WebSocketHandler {
	return &WebSocketHandler{
		server: s,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// Обработка WebSocket подключения
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка апгрейда до WebSocket: %v", err)
		return
	}
	
	clientID := uuid.New().String()
	client := client.NewClient(clientID, conn, h.server)
	h.server.Register(client)
	
	go client.WritePump()
	go client.ReadPump()
}

// Статус сервера
func (h *WebSocketHandler) StatusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":       "running",
		"client_count": h.server.GetClientCount(),
		"clients":      h.server.GetClients(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HTML страница для тестирования
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WebSocket Test Client</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px 30px;
        }
        .header h1 { font-size: 24px; margin-bottom: 5px; }
        .header p { opacity: 0.9; font-size: 14px; }
        .status-bar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 15px 30px;
            background: #f7f9fc;
            border-bottom: 1px solid #e1e8ed;
        }
        .connection-status { display: flex; align-items: center; gap: 10px; }
        .status-indicator {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            background: #e74c3c;
        }
        .status-indicator.connected { background: #2ecc71; }
        .client-id { font-family: monospace; font-size: 12px; color: #7f8c8d; }
        .controls {
            padding: 20px 30px;
            background: #f7f9fc;
            border-bottom: 1px solid #e1e8ed;
        }
        .button-group {
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
            margin-bottom: 15px;
        }
        button {
            padding: 10px 20px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.3s;
        }
        button:hover { transform: translateY(-2px); box-shadow: 0 5px 15px rgba(0,0,0,0.2); }
        button:active { transform: translateY(0); }
        .btn-connect { background: #2ecc71; color: white; }
        .btn-disconnect { background: #e74c3c; color: white; }
        .btn-text { background: #3498db; color: white; }
        .btn-binary { background: #9b59b6; color: white; }
        .input-group { display: flex; gap: 10px; }
        input {
            flex: 1;
            padding: 10px;
            border: 2px solid #e1e8ed;
            border-radius: 8px;
            font-size: 14px;
        }
        input:focus { outline: none; border-color: #3498db; }
        .messages-container { padding: 20px 30px; background: white; }
        .messages-header { margin-bottom: 15px; font-weight: 600; color: #2c3e50; }
        .messages {
            height: 400px;
            overflow-y: auto;
            border: 1px solid #e1e8ed;
            border-radius: 8px;
            padding: 15px;
            background: #fafafa;
        }
        .message {
            margin-bottom: 12px;
            padding: 10px;
            border-radius: 8px;
        }
        .message-text { background: #e3f2fd; border-left: 4px solid #2196f3; }
        .message-binary { background: #f3e5f5; border-left: 4px solid #9c27b0; }
        .message-system { background: #fff3e0; border-left: 4px solid #ff9800; font-style: italic; }
        .message-header { font-size: 12px; color: #7f8c8d; margin-bottom: 5px; }
        .message-content { font-size: 14px; word-wrap: break-word; }
        .timestamp { font-size: 10px; color: #95a5a6; margin-top: 5px; }
        .stats {
            padding: 15px 30px;
            background: #f7f9fc;
            border-top: 1px solid #e1e8ed;
            font-size: 12px;
            color: #7f8c8d;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 WebSocket Test Client</h1>
            <p>Real-time messaging with text and binary data support</p>
        </div>
        <div class="status-bar">
            <div class="connection-status">
                <div class="status-indicator" id="statusIndicator"></div>
                <span id="connectionStatus">Disconnected</span>
            </div>
            <div class="client-id">Client ID: <span id="clientId">-</span></div>
        </div>
        <div class="controls">
            <div class="button-group">
                <button class="btn-connect" onclick="connect()">🔌 Connect</button>
                <button class="btn-disconnect" onclick="disconnect()" disabled>❌ Disconnect</button>
                <button class="btn-text" onclick="sendText()" disabled>💬 Send Text</button>
                <button class="btn-binary" onclick="sendBinary()" disabled>📦 Send Binary</button>
            </div>
            <div class="input-group">
                <input type="text" id="messageInput" placeholder="Enter your message..." disabled>
            </div>
        </div>
        <div class="messages-container">
            <div class="messages-header">📨 Messages</div>
            <div class="messages" id="messages">
                <div class="message message-system">
                    <div class="message-content">👋 Welcome! Click "Connect" to start chatting.</div>
                </div>
            </div>
        </div>
        <div class="stats"><span id="stats">Ready to connect</span></div>
    </div>
    
    <script>
        let ws = null;
        let clientId = null;
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        function addMessage(type, data, timestamp) {
            const messagesDiv = document.getElementById('messages');
            const messageDiv = document.createElement('div');
            messageDiv.className = 'message message-' + type;
            
            const time = timestamp ? new Date(timestamp).toLocaleTimeString() : new Date().toLocaleTimeString();
            
            let header = '';
            let content = '';
            
            if (type === 'system') {
                header = 'System';
                content = data;
            } else if (type === 'text') {
                header = data.client_id ? data.client_id.substring(0, 8) + '...' : 'Client';
                content = data.content.text;
            } else if (type === 'binary') {
                header = data.client_id ? data.client_id.substring(0, 8) + '...' : 'Client';
                content = '📦 Binary data received';
            }
            
            const headerDiv = document.createElement('div');
            headerDiv.className = 'message-header';
            headerDiv.textContent = header;
            
            const contentDiv = document.createElement('div');
            contentDiv.className = 'message-content';
            contentDiv.textContent = content;
            
            const timeDiv = document.createElement('div');
            timeDiv.className = 'timestamp';
            timeDiv.textContent = time;
            
            messageDiv.appendChild(headerDiv);
            messageDiv.appendChild(contentDiv);
            messageDiv.appendChild(timeDiv);
            
            messagesDiv.appendChild(messageDiv);
            messageDiv.scrollIntoView({ behavior: 'smooth', block: 'end' });
            
            const messageCount = messagesDiv.children.length;
            document.getElementById('stats').textContent = 'Messages: ' + messageCount + ' | Last message: ' + time;
        }
        
        function updateUI(connected) {
            const connectBtn = document.querySelector('.btn-connect');
            const disconnectBtn = document.querySelector('.btn-disconnect');
            const sendTextBtn = document.querySelector('.btn-text');
            const sendBinaryBtn = document.querySelector('.btn-binary');
            const messageInput = document.getElementById('messageInput');
            const statusIndicator = document.getElementById('statusIndicator');
            const connectionStatus = document.getElementById('connectionStatus');
            
            if (connected) {
                connectBtn.disabled = true;
                disconnectBtn.disabled = false;
                sendTextBtn.disabled = false;
                sendBinaryBtn.disabled = false;
                messageInput.disabled = false;
                statusIndicator.classList.add('connected');
                connectionStatus.textContent = 'Connected';
            } else {
                connectBtn.disabled = false;
                disconnectBtn.disabled = true;
                sendTextBtn.disabled = true;
                sendBinaryBtn.disabled = true;
                messageInput.disabled = true;
                statusIndicator.classList.remove('connected');
                connectionStatus.textContent = 'Disconnected';
                document.getElementById('clientId').textContent = '-';
            }
        }
        
        function connect() {
            ws = new WebSocket('ws://localhost:8080/ws');
            
            ws.onopen = function() {
                console.log('WebSocket connected');
                updateUI(true);
                addMessage('system', 'Connected to server', null);
            };
            
            ws.onmessage = function(event) {
                try {
                    const data = JSON.parse(event.data);
                    console.log('Received:', data);
                    
                    if (data.client_id && data.client_id !== 'system') {
                        document.getElementById('clientId').textContent = data.client_id.substring(0, 8) + '...';
                    }
                    
                    addMessage(data.type, data, data.timestamp);
                } catch (e) {
                    console.error('Error parsing message:', e);
                }
            };
            
            ws.onerror = function(error) {
                console.error('WebSocket error:', error);
                addMessage('system', 'Connection error', null);
            };
            
            ws.onclose = function() {
                console.log('WebSocket disconnected');
                updateUI(false);
                addMessage('system', 'Disconnected from server', null);
                ws = null;
            };
        }
        
        function disconnect() {
            if (ws) {
                ws.close();
            }
        }
        
        function sendText() {
            if (ws && ws.readyState === WebSocket.OPEN) {
                const input = document.getElementById('messageInput');
                const text = input.value.trim();
                
                if (text === '') {
                    addMessage('system', 'Please enter a message', null);
                    return;
                }
                
                const message = {
                    type: 'text',
                    content: { text: text }
                };
                
                ws.send(JSON.stringify(message));
                addMessage('text', {
                    client_id: 'You',
                    content: { text: text }
                }, null);
                input.value = '';
            } else {
                addMessage('system', 'Not connected to server', null);
            }
        }
        
        function sendBinary() {
            if (ws && ws.readyState === WebSocket.OPEN) {
                const testData = new Uint8Array([72, 101, 108, 108, 111, 32, 87, 101, 98, 83, 111, 99, 107, 101, 116, 33]);
                ws.send(testData);
                addMessage('binary', {
                    client_id: 'You',
                    content: { data: 'Binary data sent', encoding: 'raw' }
                }, null);
            } else {
                addMessage('system', 'Not connected to server', null);
            }
        }
        
        document.getElementById('messageInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter' && !this.disabled) {
                sendText();
            }
        });
    </script>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}