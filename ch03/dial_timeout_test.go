package ch03

import (
	"net"
	"syscall"
	"testing"
	"time"
)

// DialTimeout은 지정된 네트워크와 주소로 지정된 시간 내에 연결을 시도하는 함수입니다.
func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	// net.Dialer 구조체를 생성합니다.
	d := net.Dialer{
		// Control 필드에 연결 시도 시 특정 제어 동작을 정의합니다.
		Control: func(_, addr string, _ syscall.RawConn) error {
			// 연결 시도 시 항상 타임아웃 에러를 반환합니다.
			return &net.DNSError{
				Err:         "connection timed out", // 에러 메시지
				Name:        addr,                  // 주소
				Server:      "127.0.0.1",           // 서버 주소
				IsTimeout:   true,                  // 타임아웃 플래그
				IsTemporary: true,                  // 임시 에러 플래그
			}
		},
		Timeout: timeout, // 연결 타임아웃을 설정합니다.
	}
	return d.Dial(network, address) // 지정된 네트워크와 주소로 연결을 시도합니다.
}

// TestDialTimeout은 DialTimeout 함수의 동작을 테스트합니다.
func TestDialTimeout(t *testing.T) {
	// DialTimeout 함수를 호출하여 10.0.0.1 주소로 5초 타임아웃을 설정합니다.
	conn, err := DialTimeout("tcp", "10.0.0.1:http", 5*time.Second)
	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out") // 타임아웃이 발생하지 않으면 실패
	}

	// 반환된 에러가 net.Error 타입인지 확인합니다.
	nErr, ok := err.(net.Error)
	if !ok {
		t.Fatal(err) // net.Error 타입이 아니면 실패
	}

	// 반환된 에러가 타임아웃 에러인지 확인합니다.
	if !nErr.Timeout() {
		t.Fatal("error is not a timeout") // 타임아웃 에러가 아니면 실패
	}
}
