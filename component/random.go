package component

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
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

// handleConnections handles WebSocket requests from clients
func HandleRandomGeneration(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	log.Println("Client connected")

	for {
		// Read message from the client
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// 초기 데이터 전송
		seed, err := generateSeed()
		if err != nil {
			fmt.Println("난수 생성 실패:", err)
			return
		}

		randomArray := generateRandomArray(seed)
		response, err := json.Marshal(randomArray)
		if err != nil {
			log.Println("Error marshaling JSON:", err)
			break
		}

		err = conn.WriteMessage(websocket.TextMessage, response)
		if err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
}
