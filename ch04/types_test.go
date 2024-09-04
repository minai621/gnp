package ch04

import (
	"bytes"
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

// TestPayloads 함수는 Binary와 String 타입의 데이터를 TCP를 통해 전송하고
// 제대로 전송 및 수신되는지 확인합니다.
func TestPayloads(t *testing.T) {
	// 테스트용 Payload 객체들을 생성
	b1 := Binary("Clear is better than clever.") // Binary 타입
	b2 := Binary("Don't panic.")                 // Binary 타입
	s1 := String("Errors are values.")           // String 타입
	payloads := []Payload{&b1, &s1, &b2}         // Payload 인터페이스 슬라이스로 묶음

	// TCP 리스너를 시작
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err) // 리스너 시작 실패 시 테스트 종료
	}

	// 리스너에서 연결을 수락하는 고루틴을 시작
	go func() {
		conn, err := listener.Accept() // 연결 수락
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()

		// 각 Payload 객체를 순차적으로 TCP 연결로 전송
		for _, p := range payloads {
			_, err = p.WriteTo(conn) // 데이터를 conn에 씀
			if err != nil {
				t.Error(err)
				break
			}
		}
	}()

	// 클라이언트 측에서 리스너로 연결을 시도
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// 수신된 데이터가 기대한 Payload와 일치하는지 확인
	for i := 0; i < len(payloads); i++ {
		actual, err := decode(conn) // conn으로부터 데이터를 읽어와 디코딩
		if err != nil {
			t.Fatal(err)
		}

		// 기대하는 값과 실제 수신된 값이 일치하는지 비교
		if expected := payloads[i]; !reflect.DeepEqual(expected, actual) {
			t.Errorf("value mismatch: %v != %v", expected, actual)
			continue
		}

		t.Logf("[%T] %[1]q", actual) // 디코딩된 값의 타입과 내용을 로그에 출력
	}
}

// TestMaxPayloadSize 함수는 최대 페이로드 크기 제한이 올바르게 작동하는지 확인합니다.
func TestMaxPayloadSize(t *testing.T) {
	// 새로운 바이트 버퍼 생성
	buf := new(bytes.Buffer)
	// 버퍼에 BinaryType 바이트 작성
	err := buf.WriteByte(BinaryType)
	if err != nil {
		t.Fatal(err)
	}

	// 버퍼에 매우 큰 크기(1GB)의 페이로드 크기 작성
	err = binary.Write(buf, binary.BigEndian, uint32(1<<30)) // 1GB 크기
	if err != nil {
		t.Fatal(err)
	}

	// Binary 타입의 객체 생성
	var b Binary
	// 버퍼로부터 데이터를 읽어서 Binary 객체에 저장
	_, err = b.ReadFrom(buf)
	// 최대 페이로드 크기 초과 에러가 발생했는지 확인
	if err != ErrMaxPayloadSize {
		t.Fatalf("expected ErrMaxPayloadSize; actual: %v", err)
	}
}
