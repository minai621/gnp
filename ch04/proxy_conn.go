package ch04

import (
	"io"  // 입출력 작업을 위한 패키지
	"net" // 네트워크 관련 기능을 제공하는 패키지
)

// proxyConn 함수는 소스와 목적지 주소를 받아서 TCP 연결을 설정하고 데이터를 전달하는 역할을 합니다.
func proxyConn(source, destination string) error {
	// 소스 주소로 TCP 연결을 생성
	connSource, err := net.Dial("tcp", source)
	if err != nil {
		return err // 연결 실패 시 에러 반환
	}
	defer connSource.Close() // 함수 종료 시 소스 연결 닫기

	// 목적지 주소로 TCP 연결을 생성
	connDest, err := net.Dial("tcp", destination)
	if err != nil {
		return err // 연결 실패 시 에러 반환
	}
	defer connDest.Close() // 함수 종료 시 목적지 연결 닫기

	// 고루틴(goroutine)을 생성하여 데이터를 소스에서 목적지로 복사
	go func() {
		_, _ = io.Copy(connDest, connSource) // connSource로부터 읽은 데이터를 connDest로 복사
		// 복사가 완료되면 고루틴은 종료됩니다.
	}()

	// 에러가 없으므로 nil 반환
	return err
}
