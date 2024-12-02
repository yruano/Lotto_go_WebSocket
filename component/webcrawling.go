package component

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gocolly/colly/v2"
	"github.com/gorilla/websocket"
)

// WebSocket 연결을 처리하는 함수
func HandleWebCrawling(w http.ResponseWriter, r *http.Request) {
	// WebSocket 연결 업그레이드
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket 연결 실패:", err)
		return
	}
	defer conn.Close()

	// 클라이언트로부터 메시지를 받으면 크롤링하고 결과를 전송하는 기능
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("메시지 읽기 실패:", err)
			break
		}

		// 메시지가 "crawl"이면 웹 크롤링 시작
		if string(msg) == "crawl" {
			data, err := crawlWebsite("https://dhlottery.co.kr/gameResult.do?method=byWin") // 원하는 웹사이트 URL로 변경
			if err != nil {
				log.Println("웹 크롤링 오류:", err)
				conn.WriteMessage(websocket.TextMessage, []byte("웹 크롤링 오류"))
				continue
			}

			// 크롤링 결과를 클라이언트로 전송
			err = conn.WriteMessage(websocket.TextMessage, []byte(data))
			if err != nil {
				log.Println("메시지 전송 실패:", err)
				break
			}
		}
	}
}

// 웹 크롤링 함수
func crawlWebsite(url string) (string, error) {
	// colly 크롤러 생성
	c := colly.NewCollector()

	// 페이지에서 원하는 데이터를 추출하는 콜백 함수
	var result string
	c.OnHTML("div.win_result", func(e *colly.HTMLElement) {
		result = e.Text // 예시로 페이지의 title을 추출
	})

	// 오류 처리
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("크롤링 오류:", err)
	})

	// 웹 페이지 크롤링
	err := c.Visit(url)
	if err != nil {
		return "", err
	}

	// 크롤링한 데이터를 반환
	return result, nil
}
