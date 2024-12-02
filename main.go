package main

import (
	"fmt"
	"net/http"

	"lotto_server/component"
)

func main() {
	// WebSocket 핸들러 등록
	http.HandleFunc("/rg", component.HandleRandomGeneration)
	http.HandleFunc("/wc", component.HandleWebCrawling)

	
  // 서버 실행
	fmt.Println("서버 실행 중: http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("서버 실행 실패:", err)
	}
}
