package ch03

import (
	"context"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

// Pinger는 주어진 시간 간격(interval)마다 "ping" 메시지를 io.Writer에 작성합니다.
// ctx가 취소되거나 reset 채널을 통해 새로운 간격이 전달될 때까지 계속 동작합니다.
func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration

	// 초기화 단계에서 컨텍스트가 완료되었거나, reset 채널에서 새로운 간격이 전달된 경우 처리
	select {
	case <-ctx.Done():  // 컨텍스트가 완료된 경우
		return
	case interval = <-reset:  // reset 채널에서 새로운 간격이 전달된 경우
	default:
	}

	// interval이 0으로 설정되면 기본 간격을 사용
	if interval == 0 {
		interval = defaultPingInterval
	}

	// 타이머를 설정하고, 함수가 끝날 때 타이머를 정리
	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {  // 타이머를 정리
			<-timer.C  // 타이머 채널을 읽어버려서 타이머의 잔여 이벤트를 소모
		}
	}()

	for {
		select {
		case <-ctx.Done():  // 컨텍스트가 완료되면 루프 종료
			return
		case newInterval := <-reset:  // reset 채널에서 새로운 간격이 전달된 경우
			if !timer.Stop() {  // 기존 타이머 정리
				<-timer.C  // 타이머 채널을 읽어버려서 타이머의 잔여 이벤트를 소모
			}
			if newInterval > 0 {
				interval = newInterval  // 새로운 간격으로 업데이트
			}
		case <-timer.C:  // 타이머가 만료되면 "ping" 메시지 작성
			if _, err := w.Write([]byte("ping\n")); err != nil {
				return  // 오류 발생 시 루프 종료
			}
		}
		_ = timer.Reset(interval)  // 타이머를 새로운 간격으로 리셋
	}
}
