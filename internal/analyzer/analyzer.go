package analyzer

import (
	"context"
	"io"
	"log"

	"github.com/google/uuid"
	mPb "github.com/orgwats/idl/gen/go/market"
	"google.golang.org/grpc"
)

type Analyzer struct {
	ctx   context.Context
	store interface{}

	client mPb.MarketClient
	stream grpc.ServerStreamingClient[mPb.Candle]

	symbol       string
	CandleBuffer *CandleBuffer

	Stop func() error
}

func NewAnalyzer(
	ctx context.Context,
	store interface{},
	client mPb.MarketClient,
	stream grpc.ServerStreamingClient[mPb.Candle],
	symbol string,
	stop func() error,
) (uuid.UUID, *Analyzer) {
	uuid, _ := uuid.NewUUID()
	analyzer := &Analyzer{
		ctx:    ctx,
		store:  store,
		client: client,
		stream: stream,
		symbol: symbol,
		CandleBuffer: &CandleBuffer{
			buf:  make([]*mPb.Candle, 200),
			head: 0,
			size: 200,
		},
		Stop: stop,
	}

	return uuid, analyzer
}

func (a *Analyzer) Init() error {
	resp, err := a.client.GetCandles(a.ctx, &mPb.GetCandlesRequest{Symbol: a.symbol, Limit: 200})
	if err != nil {
		a.Stop()
		return err
	}

	candle, err := a.stream.Recv()
	if err != nil {
		a.Stop()
		return err
	}

	a.CandleBuffer.Init(append(resp.Candles, candle))

	return nil
}

func (a *Analyzer) Run() {
	for {
		select {
		case <-a.ctx.Done():
			return
		default:
			candle, err := a.stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				// 필요에 따라 재시도 로직
				return
			}

			a.CandleBuffer.UpdateLastCandle(candle)

			log.Printf("[%s] RSI : %f", a.symbol, CalculateRSI(a.CandleBuffer.GetCandles(), 14))

			if candle.Closed {
				a.CandleBuffer.AddCandle(candle)
			}
		}
	}
}
