package ch03

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanOut(t *testing.T) {
	// 5초의 데드라인을 가진 컨텍스트 생성
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))

	// TCP 리스너 생성, 무작위 포트에서 수신 대기
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// 고루틴에서 연결 수락(accept)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	// dial 함수는 각 고루틴에서 TCP 연결을 시도
	dial := func(ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
		defer wg.Done()  // 작업이 끝나면 WaitGroup에 작업 완료를 알림

		var d net.Dialer

		// 주어진 주소로 TCP 연결 시도
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return  // 오류 발생 시 함수 종료
		}
		c.Close()  // 연결이 성공하면 닫음

		select {
		case <-ctx.Done():  // 컨텍스트가 취소되었는지 확인
		case response <- id:  // 취소되지 않았다면 응답 채널에 ID 전송
		}
	}

	// 응답을 받을 채널 및 동기화를 위한 WaitGroup 생성
	res := make(chan int)
	var wg sync.WaitGroup

	// 10개의 고루틴을 생성하여 병렬로 TCP 연결 시도
	for i := 0; i < 10; i++ {
		wg.Add(1)  // 각 고루틴이 시작되기 전에 WaitGroup에 작업 추가
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}

	// res 채널로부터 응답을 기다림, 가장 먼저 성공한 고루틴의 ID를 받음
	response := <-res
	cancel()  // 응답을 받은 후 모든 고루틴이 작업을 중지하도록 컨텍스트 취소
	wg.Wait() // 모든 고루틴이 종료될 때까지 대기
	close(res) // res 채널 닫기

	// 컨텍스트가 취소되었는지 확인
	if ctx.Err() != context.Canceled {
		t.Errorf("expected context canceled; actual %s", ctx.Err())
	}

	// 가장 먼저 성공한 Dialer의 ID를 로그에 출력
	t.Logf("dialer %d retrieved the resource", response)
}
