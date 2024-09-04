package ch04

import (
	"io"
	"net"
	"sync"
	"testing"
)

// proxy 함수는 "from"에서 "to"로 데이터를 복사하며, 필요할 경우 역방향 프록시도 수행합니다.
func proxy(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer) // from이 io.Writer 인터페이스를 구현하는지 확인
	toReader, toIsReader := to.(io.Reader)       // to가 io.Reader 인터페이스를 구현하는지 확인

	if toIsReader && fromIsWriter {
		// to가 io.Reader를 구현하고 from이 io.Writer를 구현하면
		// 역방향으로 데이터를 복사하는 고루틴을 생성
		go func() { _, _ = io.Copy(fromWriter, toReader) }()
	}

	// 기본적으로 "from"에서 "to"로 데이터를 복사
	_, err := io.Copy(to, from)
	return err
}

// TestProxy 함수는 프록시 서버가 정상적으로 작동하는지 테스트합니다.
func TestProxy(t *testing.T) {
	var wg sync.WaitGroup // 고루틴 동기화를 위한 WaitGroup

	// 서버를 설정하고 "ping" 메시지에 "pong"으로 응답, 그 외의 메시지는 에코
	server, err := net.Listen("tcp", "127.0.0.1:") // 랜덤한 포트에서 TCP 서버 생성
	if err != nil {
		t.Fatal(err) // 서버 설정 실패 시 테스트 종료
	}

	wg.Add(1) // 서버 고루틴 대기 설정
	go func() {
		defer wg.Done() // 고루틴 완료 시 WaitGroup 카운트 감소

		for {
			conn, err := server.Accept() // 클라이언트 연결 수락
			if err != nil {
				return
			}

			// 새로운 고루틴에서 클라이언트 연결 처리
			go func(c net.Conn) {
				defer c.Close()

				for {
					buf := make([]byte, 1024) // 버퍼 생성
					n, err := c.Read(buf)     // 클라이언트로부터 데이터 읽기
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}

					switch msg := string(buf[:n]); msg {
					case "ping": // "ping" 메시지에 대한 응답
						_, err = c.Write([]byte("pong"))
					default: // 다른 메시지는 에코
						_, err = c.Write(buf[:n])
					}

					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
				}
			}(conn)
		}
	}()

	// 프록시 서버 설정
	proxyServer, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err) // 프록시 서버 설정 실패 시 테스트 종료
	}

	wg.Add(1) // 프록시 서버 고루틴 대기 설정
	go func() {
		defer wg.Done() // 고루틴 완료 시 WaitGroup 카운트 감소

		for {
			conn, err := proxyServer.Accept() // 클라이언트 연결 수락
			if err != nil {
				return
			}

			// 새로운 고루틴에서 클라이언트 연결 처리
			go func(from net.Conn) {
				defer from.Close()
				to, err := net.Dial("tcp", server.Addr().String()) // 실제 서버와의 연결 설정
				if err != nil {
					t.Error(err)
					return
				}
				defer to.Close()

				// 프록시 기능 수행
				err = proxy(from, to)
				if err != nil && err != io.EOF {
					t.Error(err)
				}
			}(conn)
		}
	}()

	// 클라이언트 역할: 프록시 서버에 연결하고 메시지를 전송 및 수신
	conn, err := net.Dial("tcp", proxyServer.Addr().String()) // 프록시 서버에 연결
	if err != nil {
		t.Fatal(err)
	}

	// 테스트할 메시지와 기대 응답을 정의
	msgs := []struct{ Message, Reply string }{
		{"ping", "pong"},
		{"pong", "pong"},
		{"echo", "echo"},
		{"ping", "pong"},
	}

	// 메시지를 전송하고 기대 응답을 검증
	for i, m := range msgs {
		_, err = conn.Write([]byte(m.Message)) // 메시지 전송
		if err != nil {
			t.Fatal(err)
		}

		buf := make([]byte, 1024) // 응답을 수신할 버퍼
		n, err := conn.Read(buf)  // 응답 읽기
		if err != nil {
			t.Fatal(err)
		}

		if actual := string(buf[:n]); actual != m.Reply { // 기대한 응답과 실제 응답 비교
			t.Errorf("%d: expected reply: %q; actual: %q",
				i, m.Reply, actual)
		}
	}

	// 연결 및 리소스 해제
	_ = conn.Close()
	_ = proxyServer.Close()
	_ = server.Close()
	wg.Wait() // 모든 고루틴이 완료될 때까지 대기
}
