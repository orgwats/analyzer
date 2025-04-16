package analyzer

import (
	"context"
	"io"
	"strconv"

	"github.com/google/uuid"
	"github.com/orgwats/analyzer/internal/kafka"
	mPb "github.com/orgwats/idl/gen/go/market"
	"google.golang.org/grpc"
)

type Analyzer struct {
	Id    string
	ctx   context.Context
	store interface{}

	symbol string

	client   mPb.MarketClient
	stream   grpc.ServerStreamingClient[mPb.Candle]
	producer *kafka.Producer

	Position *Position

	CandleBuffer *CandleBuffer

	Stop func() error
}

type Position struct {
	Type       string // "LONG", "SHORT", ""
	Price      float64
	TakeProfit float64
	StopLoss   float64
	IsFilled   bool
}

func NewAnalyzer(
	ctx context.Context,
	store interface{},
	client mPb.MarketClient,
	stream grpc.ServerStreamingClient[mPb.Candle],
	producer *kafka.Producer,
	symbol string,
	stop func() error,
) (string, *Analyzer) {
	analyzer := &Analyzer{
		Id:       uuid.New().String(),
		ctx:      ctx,
		store:    store,
		client:   client,
		stream:   stream,
		producer: producer,
		symbol:   symbol,
		Position: &Position{
			Type:       "",
			Price:      0,
			TakeProfit: 0,
			StopLoss:   0,
			IsFilled:   false,
		},
		CandleBuffer: &CandleBuffer{
			buf:  make([]*mPb.Candle, 200),
			head: 0,
			size: 200,
		},
		Stop: stop,
	}

	return analyzer.Id, analyzer
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
				return
			}

			a.CandleBuffer.UpdateLastCandle(candle)

			if a.Position.Type == "" {
				orderSide := ""

				rsi := CalculateRSI(a.CandleBuffer.GetCandles(), 14)

				if rsi < 40 {
					orderSide = "BUY"
				} else if rsi > 60 {
					orderSide = "SELL"
				}

				if orderSide != "" {
					price, _ := strconv.ParseFloat(candle.Close, 64)

					if orderSide == "BUY" {
						// config 에서 받아와서 초기화 해야함.
						takeProfitROI := 1.0007
						stopLossROI := 0.9996

						a.Position.Type = "LONG"
						a.Position.Price = price
						a.Position.TakeProfit = price * takeProfitROI
						a.Position.StopLoss = price * stopLossROI
						a.Position.IsFilled = false
					}

					if orderSide == "SELL" {
						// config 에서 받아와서 초기화 해야함.
						takeProfitROI := 0.9996
						stopLossROI := 1.0007

						a.Position.Type = "SHORT"
						a.Position.Price = price
						a.Position.TakeProfit = price * takeProfitROI
						a.Position.StopLoss = price * stopLossROI
						a.Position.IsFilled = false
					}

					err = a.sendOrder(orderSide)
					if err != nil {
						// 필요에 따라 재시도 로직
					}
				}
			} else {
				price, _ := strconv.ParseFloat(candle.Close, 64)

				if a.Position.Type == "LONG" {
					if price >= a.Position.TakeProfit || price < a.Position.StopLoss {
						err := a.sendOrder("SELL")
						if err != nil {
							// 필요에 따라 재주문 로직
						}
						a.resetPosition()
					}
				}

				if a.Position.Type == "SHORT" {
					if price <= a.Position.TakeProfit || price > a.Position.StopLoss {
						err := a.sendOrder("BUY")
						if err != nil {
							// 필요에 따라 재주문 로직
						}
						a.resetPosition()
					}
				}
			}

			if candle.Closed {
				a.CandleBuffer.AddCandle(candle)
			}
		}
	}
}

func (a *Analyzer) resetPosition() {
	a.Position.Type = ""
	a.Position.Price = 0
	a.Position.TakeProfit = 0
	a.Position.StopLoss = 0
	a.Position.IsFilled = false
}

func (a *Analyzer) sendOrder(side string) error {
	order := &kafka.Order{
		OrderSide: side,
		Key:       uuid.New().String(),
		Symbol:    a.symbol,
		Size:      "0.03", // 테스트 고정 사이즈
	}

	_, _, err := a.producer.Send(a.Id, order)
	if err != nil {
		return err
	}

	return nil
}
