package ch04

import (
	"bytes"
	"encoding/binary" // 바이너리 데이터의 읽기 및 쓰기 패키지
	"errors"          // 에러 처리 패키지
	"fmt"             // 포맷 처리 패키지
	"io"              // 입출력 인터페이스 패키지
)

// 상수 정의
const (
	BinaryType uint8 = iota + 1 // 1 (Binary 타입 식별자)
	StringType                  // 2 (String 타입 식별자)

	MaxPayloadSize uint32 = 10 << 20 // 최대 페이로드 크기, 10 MB
)

// 에러 정의
var ErrMaxPayloadSize = errors.New("maximum payload size exceeded") // 최대 페이로드 크기 초과 에러

// Payload 인터페이스 정의: Stringer, ReaderFrom, WriterTo 인터페이스와 Bytes 메서드 포함
type Payload interface {
	fmt.Stringer
	io.ReaderFrom
	io.WriterTo
	Bytes() []byte
}

// Binary 타입 정의, []byte의 별칭
type Binary []byte

// Binary 타입의 Bytes 메서드 구현, 바이트 슬라이스를 반환
func (m Binary) Bytes() []byte  { return m }

// Binary 타입의 String 메서드 구현, 바이트 슬라이스를 문자열로 변환하여 반환
func (m Binary) String() string { return string(m) }

// Binary 타입의 WriteTo 메서드 구현, 데이터를 io.Writer로 씀
func (m Binary) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, BinaryType) // 타입을 1 바이트로 작성
	if err != nil {
		return 0, err // 에러 발생 시 0과 에러 반환
	}
	var n int64 = 1 // 쓴 바이트 수 초기화

	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 데이터 길이를 4 바이트로 작성
	if err != nil {
		return n, err
	}
	n += 4 // 길이 필드 크기 추가

	o, err := w.Write(m) // 실제 페이로드 데이터 작성
	return n + int64(o), err
}

// Binary 타입의 ReadFrom 메서드 구현, 데이터를 io.Reader로부터 읽음
func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) // 타입을 1 바이트로 읽음
	if err != nil {
		return 0, err
	}
	var n int64 = 1 // 읽은 바이트 수 초기화
	if typ != BinaryType {
		return n, errors.New("invalid Binary") // 타입이 맞지 않으면 에러 반환
	}

	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // 데이터 길이를 4 바이트로 읽음
	if err != nil {
		return n, err
	}
	n += 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize // 최대 페이로드 크기 초과 시 에러 반환
	}

	*m = make([]byte, size)
	o, err := r.Read(*m) // 실제 페이로드 데이터 읽기
	return n + int64(o), err
}

// String 타입 정의
type String string

// String 타입의 Bytes 메서드 구현, 바이트 슬라이스로 변환하여 반환
func (m String) Bytes() []byte  { return []byte(m) }

// String 타입의 String 메서드 구현
func (m String) String() string { return string(m) }

// String 타입의 WriteTo 메서드 구현, 데이터를 io.Writer로 씀
func (m String) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, StringType) // 타입을 1 바이트로 작성
	if err != nil {
		return 0, err
	}
	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 데이터 길이를 4 바이트로 작성
	if err != nil {
		return n, err
	}
	n += 4

	o, err := w.Write([]byte(m)) // 실제 페이로드 데이터 작성
	return n + int64(o), err
}

// String 타입의 ReadFrom 메서드 구현, 데이터를 io.Reader로부터 읽음
func (m *String) ReadFrom(r io.Reader) (int64, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) // 타입을 1 바이트로 읽음
	if err != nil {
		return 0, err
	}
	var n int64 = 1
	if typ != StringType {
		return n, errors.New("invalid String") // 타입이 맞지 않으면 에러 반환
	}

	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // 데이터 길이를 4 바이트로 읽음
	if err != nil {
		return n, err
	}
	n += 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize // 최대 페이로드 크기 초과 시 에러 반환
	}

	buf := make([]byte, size)
	o, err := r.Read(buf) // 실제 페이로드 데이터 읽기
	if err != nil {
		return n, err
	}
	*m = String(buf)

	return n + int64(o), nil
}

// decode 함수: 주어진 Reader에서 Payload를 역직렬화함
func decode(r io.Reader) (Payload, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) // 타입을 1 바이트로 읽음
	if err != nil {
		return nil, err
	}

	var payload Payload

	switch typ {
	case BinaryType:
		payload = new(Binary) // Binary 타입으로 설정
	case StringType:
		payload = new(String) // String 타입으로 설정
	default:
		return nil, errors.New("unknown type") // 알 수 없는 타입 에러
	}

	_, err = payload.ReadFrom(
		io.MultiReader(bytes.NewReader([]byte{typ}), r)) // 타입 바이트를 포함하여 페이로드 읽기
	if err != nil {
		return nil, err
	}

	return payload, nil
}
