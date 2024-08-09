package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) { 
    // 현재 시간에서 5초 후를 데드라인으로 설정
    dl := time.Now().Add(5 * time.Second)
    
    // 데드라인을 설정한 컨텍스트 생성
    ctx, cancel := context.WithDeadline(context.Background(), dl)
    
    // 함수가 종료되면 cancel 함수 호출
    // 이는 컨텍스트와 연관된 리소스를 해제하고, 타임아웃이나 취소를 발생시키는 데 사용
    defer cancel()

    var d net.Dialer

    // Dialer의 Control 함수 정의
    // 이 함수는 네트워크 연결이 설정되기 전에 호출되며, 여기서는 5초 + 1밀리초의 인위적인 지연을 추가
    d.Control = func(_, _ string, c syscall.RawConn) error {
        time.Sleep(5 * time.Second + time.Millisecond)
        return nil
    }

    // TCP 연결을 시도하며, 컨텍스트를 통해 타임아웃 제어
    conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
    if err != nil {
        // 오류가 발생한 경우, 연결 객체가 nil이 아니면 Close 호출하여 자원 해제
        if conn != nil {
            conn.Close()
        }
        t.Fatal("connection did not time out") // 타임아웃이 발생하지 않았다면 테스트 실패
    }

    // 오류가 net.Error 타입인지 확인
    nErr, ok := err.(net.Error)
    if !ok {
        t.Error(err) // net.Error 타입이 아닌 경우 테스트 실패
    } else {
        // 오류가 타임아웃에 의한 것인지 확인
        if !nErr.Timeout() {
            t.Errorf("error is not a timeout: %v", err) // 타임아웃이 아니면 테스트 실패
        }
    }

    // 컨텍스트의 오류가 데드라인 초과인지 확인
    if ctx.Err() != context.DeadlineExceeded {
        t.Errorf("expected deadline exceeded; actual %v", ctx.Err()) // 데드라인 초과가 아니면 테스트 실패
    }
}
