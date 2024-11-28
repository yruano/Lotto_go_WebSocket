package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"lotto_server/component"
)

func main() {
	// WebSocket 핸들러 등록
	http.HandleFunc("/ws", component.HandleWebSocket)

	// 브로드캐스트 및 데이터 갱신 고루틴 실행
	go component.BroadcastUpdates()
	go component.HandleBroadcast()

	// 종료 신호 처리
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// 서버 종료를 처리하는 고루틴
	go func() {
		sig := <-quit
		fmt.Println("서버 종료 신호 받음:", sig)

		// 클라이언트 연결 종료 처리
		fmt.Println("모든 클라이언트 연결 종료 중...")
    component.HandleShutdown()

		// 서버 종료 전에 처리해야 할 작업이 있으면 여기에 추가
		fmt.Println("서버 종료 작업 완료.")
		os.Exit(0)
	}()

	// 서버 실행
	fmt.Println("서버 실행 중: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("서버 실행 실패:", err)
	}
}
