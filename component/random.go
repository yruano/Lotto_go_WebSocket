package component

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"fmt"
	mathRand "math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const updateInterval = 10 * time.Second // 예: 10초로 설정

// WebSocket 업그레이더 설정
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 요청 허용 (보안을 위해 도메인 확인 필요)
	},
}

// 클라이언트 관리 및 데이터 채널
var (
	clients   = make(map[*websocket.Conn]bool) // 연결된 클라이언트 목록
	broadcast = make(chan []int, 10)           // 브로드캐스트 데이터 채널
)

// 난수 시드 생성 함수
func generateSeed() (uint64, error) {
	nanoTime := uint64(time.Now().UnixNano())

	var randBytes [8]byte
	_, err := cryptoRand.Read(randBytes[:])
	if err != nil {
		fmt.Println("cryptoRand 실패, 기본 시드 사용:", err)
		return nanoTime, nil // 기본 시드 반환
	}

	cryptoRandValue := binary.LittleEndian.Uint64(randBytes[:])
	return nanoTime ^ cryptoRandValue, nil
}

// 시드 기반으로 난수 배열 생성
func generateRandomArray(seed uint64) []int {
	// 독립적인 난수 생성기 생성
	rng := mathRand.New(mathRand.NewSource(int64(seed)))
	randomArray := make([]int, 6)
	for i := 0; i < 6; i++ {
		randomArray[i] = rng.Intn(10) // 0~9 난수 생성
	}
	return randomArray
}

// WebSocket 연결 처리
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket 업그레이드 실패:", err)
		return
	}

	clients[conn] = true
	fmt.Println("클라이언트 연결:", conn.RemoteAddr())

	defer func() {
		conn.Close()
		delete(clients, conn)
		fmt.Println("클라이언트 연결 종료:", conn.RemoteAddr())
	}()

	// 초기 데이터 전송
	seed, err := generateSeed()
	if err != nil {
		fmt.Println("난수 생성 실패:", err)
		return
	}

	randomArray := generateRandomArray(seed)
	message := fmt.Sprintf("%v", randomArray)
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		fmt.Println("초기 데이터 전송 실패:", err)
		return
	}

	// 메시지 수신 처리
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("메시지 읽기 실패:", err)
			break
		}
	}
}

// 브로드캐스트 데이터 갱신
func BroadcastUpdates() {
	for {
		// 새로운 시드와 배열 생성
		seed, err := generateSeed()
		if err != nil {
			fmt.Println("데이터 생성 실패:", err)
			continue
		}

		randomArray := generateRandomArray(seed)
		broadcast <- randomArray
		time.Sleep(updateInterval)
	}
}

// 클라이언트에게 데이터 브로드캐스트
func HandleBroadcast() {
	for randomArray := range broadcast {
		message := fmt.Sprintf("%v", randomArray) // 배열을 문자열로 변환
		for client := range clients {
			if client != nil {
				if err := client.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					fmt.Println("데이터 전송 실패:", err)
					client.Close()
					delete(clients, client) // 연결 제거
				}
			}
		}
	}
}

func HandleShutdown() {
  for client := range clients {
    client.Close()
    delete(clients, client)
  }
}