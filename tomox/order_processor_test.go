package tomox

import (
	"github.com/tomochain/tomochain/common"
	"github.com/tomochain/tomochain/ethdb"
	"github.com/tomochain/tomochain/tomox/tradingstate"
	"math/big"
	"reflect"
	"testing"
)

func Test_getCancelFee(t *testing.T) {
	tomox := New(&DefaultConfig)
	db, _ := ethdb.NewMemDatabase()
	stateCache := tradingstate.NewDatabase(db)
	tradingStateDb, _ := tradingstate.New(common.Hash{}, stateCache)

	testTokenA := common.HexToAddress("0x0000000000000000000000000000000000000002")
	testTokenB := common.HexToAddress("0x0000000000000000000000000000000000000003")
	// set decimal
	// tokenA has decimal 10^18
	tomox.SetTokenDecimal(testTokenA, common.BasePrice)
	// tokenB has decimal 10^8
	tomox.SetTokenDecimal(testTokenB, new(big.Int).Exp(big.NewInt(10), big.NewInt(8), nil))

	// set tokenAPrice = 1 TOMO
	tradingStateDb.SetMediumPriceBeforeEpoch(tradingstate.GetTradingOrderBookHash(testTokenA, common.HexToAddress(common.TomoNativeAddress)), common.BasePrice)
	// set tokenBPrice = 1 tokenA
	tradingStateDb.SetMediumPriceBeforeEpoch(tradingstate.GetTradingOrderBookHash(testTokenB, testTokenA), common.BasePrice)

	type CancelFeeArg struct {
		feeRate *big.Int
		order   *tradingstate.OrderItem
	}
	tests := []struct {
		name string
		args CancelFeeArg
		want *big.Int
	}{

		// BASE: testTokenA,
		// QUOTE: TOMO

		// zero fee test: SELL
		{
			"TokenA/TOMO zero fee test: SELL",
			CancelFeeArg{
				feeRate: common.Big0,
				order: &tradingstate.OrderItem{
					BaseToken:  testTokenA,
					QuoteToken: common.HexToAddress(common.TomoNativeAddress),
					Quantity:   new(big.Int).SetUint64(10000),
					Side:       tradingstate.Ask,
				},
			},
			common.Big0,
		},

		// zero fee test: BUY
		{
			"TokenA/TOMO zero fee test: BUY",
			CancelFeeArg{
				feeRate: common.Big0,
				order: &tradingstate.OrderItem{
					BaseToken:  testTokenA,
					QuoteToken: common.HexToAddress(common.TomoNativeAddress),
					Quantity:   new(big.Int).SetUint64(10000),
					Side:       tradingstate.Bid,
				},
			},
			common.Big0,
		},

		// test getCancelFee: SELL
		{
			"TokenA/TOMO test getCancelFee:: SELL",
			CancelFeeArg{
				feeRate: new(big.Int).SetUint64(10), // 10/1000 = 0.1%
				order: &tradingstate.OrderItem{
					BaseToken:  common.HexToAddress(common.TomoNativeAddress),
					QuoteToken: testTokenA,
					Quantity:   new(big.Int).SetUint64(10000),
					Side:       tradingstate.Ask,
				},
			},
			common.RelayerCancelFee,
		},

		// test getCancelFee:: BUY
		{
			"TokenA/TOMO test getCancelFee:: BUY",
			CancelFeeArg{
				feeRate: new(big.Int).SetUint64(10), // 10/1000 = 0.1%
				order: &tradingstate.OrderItem{
					Quantity:   new(big.Int).SetUint64(10000),
					BaseToken:  common.HexToAddress(common.TomoNativeAddress),
					QuoteToken: testTokenA,
					Side:       tradingstate.Bid,
				},
			},
			common.RelayerCancelFee,
		},

		// BASE: TOMO
		// QUOTE: testTokenA
		// zero fee test: SELL
		{
			"TOMO/TokenA zero fee test: SELL",
			CancelFeeArg{
				feeRate: common.Big0,
				order: &tradingstate.OrderItem{
					BaseToken:  common.HexToAddress(common.TomoNativeAddress),
					QuoteToken: testTokenA,
					Quantity:   new(big.Int).SetUint64(10000),
					Side:       tradingstate.Ask,
				},
			},
			common.Big0,
		},

		// zero fee test: BUY
		{
			"TOMO/TokenA zero fee test: BUY",
			CancelFeeArg{
				feeRate: common.Big0,
				order: &tradingstate.OrderItem{
					BaseToken:  common.HexToAddress(common.TomoNativeAddress),
					QuoteToken: testTokenA,
					Quantity:   new(big.Int).SetUint64(10000),
					Side:       tradingstate.Bid,
				},
			},
			common.Big0,
		},

		// test getCancelFee: SELL
		{
			"TOMO/TokenA test getCancelFee:: SELL",
			CancelFeeArg{
				feeRate: new(big.Int).SetUint64(10), // 10/1000 = 0.1%
				order: &tradingstate.OrderItem{
					BaseToken:  common.HexToAddress(common.TomoNativeAddress),
					QuoteToken: testTokenA,
					Quantity:   new(big.Int).SetUint64(10000),
					Side:       tradingstate.Ask,
				},
			},
			common.RelayerCancelFee,
		},

		// test getCancelFee:: BUY
		{
			"TOMO/TokenA test getCancelFee:: BUY",
			CancelFeeArg{
				feeRate: new(big.Int).SetUint64(10), // 10/1000 = 0.1%
				order: &tradingstate.OrderItem{
					Quantity:   new(big.Int).SetUint64(10000),
					BaseToken:  common.HexToAddress(common.TomoNativeAddress),
					QuoteToken: testTokenA,
					Side:       tradingstate.Bid,
				},
			},
			common.RelayerCancelFee,
		},

		// BASE: testTokenB
		// QUOTE: testTokenA
		// zero fee test: SELL
		{
			"TokenB/TokenA zero fee test: SELL",
			CancelFeeArg{
				feeRate: common.Big0,
				order: &tradingstate.OrderItem{
					BaseToken:  testTokenB,
					QuoteToken: testTokenA,
					Quantity:   new(big.Int).SetUint64(10000),
					Side:       tradingstate.Ask,
				},
			},
			common.Big0,
		},

		// zero fee test: BUY
		{
			"TokenB/TokenA zero fee test: BUY",
			CancelFeeArg{
				feeRate: common.Big0,
				order: &tradingstate.OrderItem{
					BaseToken:  testTokenB,
					QuoteToken: testTokenA,
					Quantity:   new(big.Int).SetUint64(10000),
					Side:       tradingstate.Bid,
				},
			},
			common.Big0,
		},

		// test getCancelFee: SELL
		{
			"TokenB/TokenA test getCancelFee:: SELL",
			CancelFeeArg{
				feeRate: new(big.Int).SetUint64(10), // 10/1000 = 0.1%
				order: &tradingstate.OrderItem{
					BaseToken:  testTokenB,
					QuoteToken: testTokenA,
					Quantity:   new(big.Int).SetUint64(10000),
					Side:       tradingstate.Ask,
				},
			},
			new(big.Int).Exp(big.NewInt(10), big.NewInt(4), nil),
		},

		// test getCancelFee:: BUY
		{
			"TokenB/TokenA test getCancelFee:: BUY",
			CancelFeeArg{
				feeRate: new(big.Int).SetUint64(10), // 10/1000 = 0.1%
				order: &tradingstate.OrderItem{
					Quantity:   new(big.Int).SetUint64(10000),
					BaseToken:  testTokenB,
					QuoteToken: testTokenA,
					Side:       tradingstate.Bid,
				},
			},
			common.RelayerCancelFee,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tomox.getCancelFee(nil, nil, tradingStateDb, tt.args.order, tt.args.feeRate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCancelFee() = %v, quantity %v", got, tt.want)
			}
		})
	}
}

func TestGetTradeQuantity(t *testing.T) {
	type GetTradeQuantityArg struct {
		takerSide        string
		takerFeeRate     *big.Int
		takerBalance     *big.Int
		makerPrice       *big.Int
		makerFeeRate     *big.Int
		makerBalance     *big.Int
		baseTokenDecimal *big.Int
		quantityToTrade  *big.Int
	}
	tests := []struct {
		name        string
		args        GetTradeQuantityArg
		quantity    *big.Int
		rejectMaker bool
	}{
		{
			"BUY: feeRate = 0, price 1, quantity 1000, taker balance 1000, maker balance 1000",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Bid,
				takerFeeRate:     common.Big0,
				takerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			false,
		},
		{
			"BUY: feeRate = 0, price 1, quantity 1000, taker balance 1000, maker balance 900 -> reject maker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Bid,
				takerFeeRate:     common.Big0,
				takerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(900), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			new(big.Int).Mul(big.NewInt(900), common.BasePrice),
			true,
		},
		{
			"BUY: feeRate = 0, price 1, quantity 1000, taker balance 900, maker balance 1000 -> reject taker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Bid,
				takerFeeRate:     common.Big0,
				takerBalance:     new(big.Int).Mul(big.NewInt(900), common.BasePrice),
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			new(big.Int).Mul(big.NewInt(900), common.BasePrice),
			false,
		},
		{
			"BUY: feeRate = 0, price 1, quantity 1000, taker balance 0, maker balance 1000 -> reject taker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Bid,
				takerFeeRate:     common.Big0,
				takerBalance:     common.Big0,
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			common.Big0,
			false,
		},
		{
			"BUY: feeRate = 0, price 1, quantity 1000, taker balance 0, maker balance 0 -> reject both taker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Bid,
				takerFeeRate:     common.Big0,
				takerBalance:     common.Big0,
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     common.Big0,
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			common.Big0,
			false,
		},
		{
			"BUY: feeRate = 0, price 1, quantity 1000, taker balance 500, maker balance 100 -> reject both taker, maker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Bid,
				takerFeeRate:     common.Big0,
				takerBalance:     new(big.Int).Mul(big.NewInt(500), common.BasePrice),
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(100), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			new(big.Int).Mul(big.NewInt(100), common.BasePrice),
			true,
		},

		{
			"SELL: feeRate = 0, price 1, quantity 1000, taker balance 1000, maker balance 1000",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Ask,
				takerFeeRate:     common.Big0,
				takerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			false,
		},
		{
			"SELL: feeRate = 0, price 1, quantity 1000, taker balance 1000, maker balance 900 -> reject maker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Ask,
				takerFeeRate:     common.Big0,
				takerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(900), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			new(big.Int).Mul(big.NewInt(900), common.BasePrice),
			true,
		},
		{
			"SELL: feeRate = 0, price 1, quantity 1000, taker balance 900, maker balance 1000 -> reject taker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Ask,
				takerFeeRate:     common.Big0,
				takerBalance:     new(big.Int).Mul(big.NewInt(900), common.BasePrice),
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			new(big.Int).Mul(big.NewInt(900), common.BasePrice),
			false,
		},
		{
			"SELL: feeRate = 0, price 1, quantity 1000, taker balance 0, maker balance 1000 -> reject taker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Ask,
				takerFeeRate:     common.Big0,
				takerBalance:     common.Big0,
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			common.Big0,
			false,
		},
		{
			"SELL: feeRate = 0, price 1, quantity 1000, taker balance 0, maker balance 0 -> reject maker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Ask,
				takerFeeRate:     common.Big0,
				takerBalance:     common.Big0,
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     common.Big0,
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			common.Big0,
			true,
		},
		{
			"SELL: feeRate = 0, price 1, quantity 1000, taker balance 500, maker balance 100 -> reject both taker, maker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Ask,
				takerFeeRate:     common.Big0,
				takerBalance:     new(big.Int).Mul(big.NewInt(500), common.BasePrice),
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(100), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			new(big.Int).Mul(big.NewInt(100), common.BasePrice),
			true,
		},
		{
			"SELL: feeRate = 0, price 1, quantity 1000, taker balance 0, maker balance 100 -> reject both taker, maker",
			GetTradeQuantityArg{
				takerSide:        tradingstate.Ask,
				takerFeeRate:     common.Big0,
				takerBalance:     common.Big0,
				makerPrice:       common.BasePrice,
				makerFeeRate:     common.Big0,
				makerBalance:     new(big.Int).Mul(big.NewInt(100), common.BasePrice),
				baseTokenDecimal: common.BasePrice,
				quantityToTrade:  new(big.Int).Mul(big.NewInt(1000), common.BasePrice),
			},
			common.Big0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetTradeQuantity(tt.args.takerSide, tt.args.takerFeeRate, tt.args.takerBalance, tt.args.makerPrice, tt.args.makerFeeRate, tt.args.makerBalance, tt.args.baseTokenDecimal, tt.args.quantityToTrade)
			if !reflect.DeepEqual(got, tt.quantity) {
				t.Errorf("GetTradeQuantity() got = %v, quantity %v", got, tt.quantity)
			}
			if got1 != tt.rejectMaker {
				t.Errorf("GetTradeQuantity() got1 = %v, quantity %v", got1, tt.rejectMaker)
			}
		})
	}
}
