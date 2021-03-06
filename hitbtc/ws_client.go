package hitbtc

import (
	"errors"

	"encoding/json"
	"strconv"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/websocket"
)

type tradeInfo struct {
	//{"id":360841011,"price":"0.031089","quantity":"0.977","side":"sell","timestamp":"2018-09-10T12:42:12.905Z"}
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
	Side      string `json:"side"` //buy sell
	Timestamp string `json:"timestamp"`
}

type paramsEntry struct {
	Data   []tradeInfo `json:"data"`
	Symbol string      `json:"symbol"`
}

type TradeDetail struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  paramsEntry `json:"params"`
}

func HitbtcWsConnect(symbolList []string) {

	if len(symbolList) <= 0 {
		logrus.Error("订阅交易对数量为空")
		return
	}

	ws, err := websocket.Dial(HitbtcWsUrl, "", HitbtcWsUrl)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	for _, s := range symbolList {

		subStr := "{\"method\": \"subscribeTrades\", \"params\":{\"symbol\": \"" + s + "\"} ,\"id\": 123}"

		_, err = ws.Write([]byte(subStr))
		if err != nil {
			logrus.Error(err.Error())
			return
		}
		logrus.Infof("订阅: %s \n", subStr)
	}

	//统计连续错误次数
	var readErrCount = 0

	var msg = make([]byte, HitbtcBufferSize)

	for {

		var data string

		for {
			if readErrCount > HitbtcErrorLimt {
				//异常退出
				ws.Close()
				logrus.Panic(errors.New("WebSocket异常连接数连续大于" + strconv.Itoa(readErrCount)))
			}

			m, err := ws.Read(msg)
			if err != nil {
				logrus.Error(err.Error())
				readErrCount++
				continue
			}
			data += string(msg[:m])
			if m <= (HitbtcBufferSize - 1) {
				break
			}
		}
		//连接正常重置
		readErrCount = 0

		logrus.Infof("Hitbtc接收：%s \n", data)
		var t TradeDetail
		err = json.Unmarshal([]byte(data), &t)
		if err != nil {
			logrus.Errorln(err)
			continue
		}
		logrus.Info("Hitbtc对象输出", t)
	}
}
