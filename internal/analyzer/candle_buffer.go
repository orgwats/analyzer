package analyzer

import (
	mPb "github.com/orgwats/idl/gen/go/market"
)

type CandleBuffer struct {
	buf  []*mPb.Candle
	head int
	size int
}

func NewCandleBuffer(size int) *CandleBuffer {
	return &CandleBuffer{
		buf:  make([]*mPb.Candle, size),
		head: 0,
		size: size,
	}
}

func (cb *CandleBuffer) Init(candles []*mPb.Candle) {
	for _, candle := range candles {
		cb.AddCandle(candle)
	}
}

func (cb *CandleBuffer) GetCandles() []*mPb.Candle {
	candles := make([]*mPb.Candle, 0, cb.size)
	for i := 0; i < cb.size; i++ {
		idx := (cb.head + i) % cb.size
		candles = append(candles, cb.buf[idx])
	}
	return candles
}

func (cb *CandleBuffer) AddCandle(candle *mPb.Candle) {
	cb.buf[cb.head] = candle
	cb.head = (cb.head + 1) % cb.size
}

func (cb *CandleBuffer) UpdateLastCandle(candle *mPb.Candle) {
	idx := (cb.head - 1 + cb.size) % cb.size
	cb.buf[idx] = candle
}
