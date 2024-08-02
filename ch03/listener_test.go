package ch03

import (
	"net"
	"testing"
)

func TestListener(t *testing.T){
	// 인터페이스와 에러 인터페이스를 반환한다. 
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	// 함수가 끝나면 listener를 닫는다.
	defer func() { _= listener.Close() } ()

	t.Logf("bound to %q", listener.Addr())
}

