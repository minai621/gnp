package ch03

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContextCancel(t *testing.T) {
	// context.Background()를 기반으로 취소 가능한 컨텍스트 생성
	ctx, cancel := context.WithCancel(context.Background())
	// 고루틴 간 동기화를 위해 빈 구조체 채널 생성
	sync := make(chan struct{})

	// 고루틴 실행
	go func() {
		defer func() {
			// 고루틴이 종료되면 sync 채널에 신호를 보냄
			sync <- struct{}{}
		}()

		// Dialer 객체 생성
		var d net.Dialer
		
		// Control 함수 정의: 이 함수는 네트워크 연결 설정 전에 호출됨
		// 여기서는 5초 + 1밀리초의 인위적인 지연을 추가
		d.Control = func(_, _ string, c syscall.RawConn) error {
			time.Sleep(5 * time.Second + time.Millisecond)
			return nil
		}

		// 지정된 주소로 TCP 연결 시도, 컨텍스트(ctx)를 통해 타임아웃 제어
		conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")
		if err != nil {
			// 오류가 발생하면 로그에 기록하고 고루틴 종료
			t.Log(err)
			return
		}
		
		// 연결이 성공하면 conn을 닫고 테스트 실패 처리 (타임아웃이 발생하지 않았기 때문)
		conn.Close()
		t.Error("connection did not time out")
	}()

	// DialContext 호출 전에 컨텍스트를 취소하여 타임아웃 발생 유도
	cancel()
	
	// sync 채널에서 신호를 받아 고루틴이 종료될 때까지 대기
	<-sync

	// 컨텍스트가 취소 상태인지 확인
	if ctx.Err() != context.Canceled {
		t.Errorf("expected context canceled; actual %v", ctx.Err())
	}
}
