package ch03

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	// 랜덤 포트에 리스너를 생성합니다.
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err) // 리스너 생성에 실패하면 테스트를 종료합니다.
	}

	// 고루틴 동기화를 위한 채널을 생성합니다.
	done := make(chan struct{})

	// 서버 역할을 하는 고루틴을 시작합니다.
	go func() {
		defer func() {
			done <- struct{}{} // 고루틴이 종료되었음을 알립니다.
		}()

		for {
			// 클라이언트 연결을 기다립니다.
			conn, err := listener.Accept()
			if err != nil {
				t.Log(err) // 연결 수락에 실패하면 에러를 로그에 기록하고 종료합니다.
				return
			}

			// 클라이언트 연결을 처리하는 또 다른 고루틴을 시작합니다.
			go func(c net.Conn) {
				defer func() {
					c.Close()          // 연결을 닫습니다.
					done <- struct{}{} // 연결 처리가 끝났음을 알립니다.
				}()

				// 버퍼를 생성하여 데이터를 읽습니다.
				buf := make([]byte, 1024)
				for {
					// 데이터를 읽습니다.
					n, err := conn.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err) // 읽기 중 에러가 발생하면 에러를 기록합니다.
						}
						return
					}
					t.Logf("received: %q", buf[:n]) // 받은 데이터를 로그에 출력합니다.
				}
			}(conn)
		}
	}()

	// 클라이언트 역할을 하는 연결을 생성합니다.
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err) // 연결 생성에 실패하면 테스트를 종료합니다.
	}

	// 클라이언트 연결을 닫습니다.
	conn.Close()
	<-done      // 첫 번째 고루틴이 종료되기를 기다립니다.
	listener.Close() // 리스너를 닫습니다.
	<-done      // 두 번째 고루틴이 종료되기를 기다립니다.
}
