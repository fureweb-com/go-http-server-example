package main

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/kataras/iris"
	"github.com/kataras/iris/websocket"
)

func main() {
	app := iris.New()

	app.Get("/", func(ctx iris.Context) {
		ctx.ServeFile("websockets.html", false) // second parameter: enable gzip?
	})

	setupWebsocket(app)

	app.Run(iris.Addr(":8080"))
}

func setupWebsocket(app *iris.Application) {
	// create our echo websocket server
	ws := websocket.New(websocket.Config{
		MaxMessageSize:  4096, // 크기 초과하는 경우 클라이언트 접속 강제 종료됨
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	})
	ws.OnConnection(handleConnection)

	app.Get("/chat", ws.Handler())

	app.Any("/ws.js", websocket.ClientHandler())
}

// SocketMessage is..
type SocketMessage struct {
	UserID   string
	Nickname string
	Message  string
	Time     int
}

func handleConnection(c websocket.Connection) {
	// Read events from browser
	c.On("chat", func(request string) {
		// ip가 필요한 경우 꺼내 사용한다.
		// context := c.Context()
		// ip := context.RemoteAddr()

		// 요청에 담긴 json string으로부터 socketMessage JSON 생성
		var socketMessage SocketMessage
		json.Unmarshal([]byte(request), &socketMessage)

		if socketMessage.UserID == "" {
			log.Println("UserId가 없음")
			return
		}

		msg, includeBadWords := filterBadWords(socketMessage.Message)

		// 자신 포함 모든 사용자에게 메시지 전달 (클라이언트에서 욕설 처리하지 않음)
		c.To(websocket.All).Emit("chat", socketMessage.Nickname+": "+msg)

		if includeBadWords == true {
			c.Emit("chat", "관리자: 욕설을 사용하는 경우 영구적으로 접속이 제한될 수 있습니다.")
		}

		// Write message back to the client message owner with: (발송자에게만 전송)
		// c.Emit("chat", msg)
		// Write message to all except this client with: (발송자 제외 전송)
		// c.To(websocket.Broadcast).Emit("chat", ip+":"+msg)
	})
}

var (
	badWords         = [...]string{"새끼", "소새끼", "말새끼"}
	badWordsReplacer = strings.NewReplacer("새끼", "응애~", "소새끼", "음메~", "말새끼", "히힝~")
)

func filterBadWords(message string) (string, bool) {
	var includeBadWords = false

	for i := 0; i < len(badWords); i++ {
		if strings.Index(message, badWords[i]) > -1 {
			includeBadWords = true
			// 비속어 사전에 있는 문자열은 모두 대체시킨다.
			message = badWordsReplacer.Replace(message)
		}
	}

	return message, includeBadWords
}
