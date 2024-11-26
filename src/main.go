package main

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"fmt"
	mathRand "math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket 업그레이더 설정
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 모든 요청 허용 (보안을 위해 도메인 확인 필요)
	},
}

// 클라이언트 관리 및 데이터 채널
var (
	clients   = make(map[*websocket.Conn]bool) // 연결된 클라이언트 목록
	broadcast = make(chan []int)               // 브로드캐스트 데이터 채널
)

// 난수 시드 생성 함수
func generateSeed() (uint64, error) {
	// 현재 시간을 기준으로 초기값 설정
	nanoTime := uint64(time.Now().UnixNano())

	// Crypto 랜덤 바이트 추가
	var randBytes [8]byte
	_, err := cryptoRand.Read(randBytes[:])
	if err != nil {
		return 0, err
	}
	cryptoRandValue := binary.LittleEndian.Uint64(randBytes[:])

	// XOR 연산으로 고유 시드 생성
	seed := nanoTime ^ cryptoRandValue

	return seed, nil
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
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket 업그레이드 실패:", err)
		return
	}
	defer conn.Close()

	// 클라이언트 등록
	clients[conn] = true
	fmt.Println("클라이언트 연결:", conn.RemoteAddr())

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
		conn.Close()
		delete(clients, conn)
		return
	}

	// 클라이언트 메시지 읽기
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("클라이언트 연결 종료:", err)
			delete(clients, conn)
			break
		}
	}
}

// 브로드캐스트 데이터 갱신
func broadcastUpdates() {
	for {
		// 새로운 시드와 배열 생성
		seed, err := generateSeed()
		if err != nil {
			fmt.Println("데이터 생성 실패:", err)
			continue
		}

		randomArray := generateRandomArray(seed)
		broadcast <- randomArray
		time.Sleep(600 * time.Second) // 5초마다 데이터 갱신
	}
}

// 클라이언트에게 데이터 브로드캐스트
func handleBroadcast() {
	for randomArray := range broadcast {
		message := fmt.Sprintf("%v", randomArray) // 배열을 문자열로 변환
		for client := range clients {
			if client != nil {
				if err := client.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
					fmt.Println("데이터 전송 실패:", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}

// 메인 함수
func main() {
	// WebSocket 핸들러 등록
	http.HandleFunc("/ws", handleWebSocket)

	// 브로드캐스트 및 데이터 갱신 고루틴 실행
	go broadcastUpdates()
	go handleBroadcast()

	// 서버 실행
	fmt.Println("서버 실행 중: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic("서버 실행 실패: " + err.Error())
	}
}
