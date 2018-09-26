package httpwebsocket

import (
	. "UNetwork/common"
	. "UNetwork/common/config"
	"UNetwork/core/ledger"
	"UNetwork/events"
	."UNetwork/net/httpjsonrpc"
	"UNetwork/net/httpwebsocket/websocket"
	"UNetwork/net/message"
	. "UNetwork/net/protocol"
	"bytes"
	"UNetwork/errors"
)

var ws *websocket.WsServer
var (
	pushBlockFlag    bool = false
	pushRawBlockFlag bool = false
	pushBlockTxsFlag bool = false
)

func StartServer(n UNode) {
	//common.SetNode(n)
	ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventBlockPersistCompleted, SendBlock2WSclient)
	ledger.DefaultLedger.Blockchain.BCEvents.Subscribe(events.EventChatMessage, SendChatMessage2WSclient)
	go func() {
		ws = websocket.InitWsServer(CheckAccessToken)
		ws.Start()
	}()
}

func SendChatMessage2WSclient(v interface{}) {
	if Parameters.HttpWsPort != 0 {
		go func() {
			PushChatMessage(v)
		}()
	}
}

func SendBlock2WSclient(v interface{}) {
	if Parameters.HttpWsPort != 0 && pushBlockFlag {
		go func() {
			PushBlock(v)
		}()
	}
	if Parameters.HttpWsPort != 0 && pushBlockTxsFlag {
		go func() {
			PushBlockTransactions(v)
		}()
	}
}
func Stop() {
	if ws == nil {
		return
	}
	ws.Stop()
}
func ReStartServer() {
	if ws == nil {
		ws = websocket.InitWsServer(CheckAccessToken)
		ws.Start()
		return
	}
	ws.Restart()
}
func GetWsPushBlockFlag() bool {
	return pushBlockFlag
}
func SetWsPushBlockFlag(b bool) {
	pushBlockFlag = b
}
func GetPushRawBlockFlag() bool {
	return pushRawBlockFlag
}
func SetPushRawBlockFlag(b bool) {
	pushRawBlockFlag = b
}
func GetPushBlockTxsFlag() bool {
	return pushBlockTxsFlag
}
func SetPushBlockTxsFlag(b bool) {
	pushBlockTxsFlag = b
}
func SetTxHashMap(txhash string, sessionid string) {
	if ws == nil {
		return
	}
	ws.SetTxHashMap(txhash, sessionid)
}

func PushResult(txHash Uint256, errcode int64, action string, result interface{}) {
	if ws != nil {
		resp := ResponsePack(errors.SUCCESS)
		resp["Result"] = result
		resp["Error"] = errcode
		resp["Action"] = action
		resp["Desc"] = errors.ErrMap[resp["Error"].(int64)]
		ws.PushTxResult(BytesToHexString(txHash.ToArrayReverse()), resp)
	}
}

func PushSmartCodeInvokeResult(txHash Uint256, errcode int64, result interface{}) {
	if ws == nil {
		return
	}
	resp := ResponsePack(errors.SUCCESS)
	var Result = make(map[string]interface{})
	txHashStr := BytesToHexString(txHash.ToArray())
	Result["TxHash"] = txHashStr
	Result["ExecResult"] = result

	resp["Result"] = Result
	resp["Action"] = "sendsmartcodeinvoke"
	resp["Error"] = errcode
	resp["Desc"] = errors.ErrMap[errcode]
	ws.PushTxResult(txHashStr, resp)
}
func PushBlock(v interface{}) {
	if ws == nil {
		return
	}
	resp := ResponsePack(errors.SUCCESS)
	if block, ok := v.(*ledger.Block); ok {
		if pushRawBlockFlag {
			w := bytes.NewBuffer(nil)
			block.Serialize(w)
			resp["Result"] = BytesToHexString(w.Bytes())
		} else {
			resp["Result"] = GetBlockInfo(block)
		}
		resp["Action"] = "sendrawblock"
		ws.PushResult(resp)
	}
}
func PushBlockTransactions(v interface{}) {
	if ws == nil {
		return
	}
	resp := ResponsePack(errors.SUCCESS)
	if block, ok := v.(*ledger.Block); ok {
		if pushBlockTxsFlag {
			resp["Result"] = GetBlockTransactions(block)
		}
		resp["Action"] = "sendblocktransactions"
		ws.PushResult(resp)
	}
}

func PushChatMessage(v interface{}) {
	if ws == nil {
		return
	}
	resp := ResponsePack(errors.SUCCESS)
	if chatMessage, ok := v.(*message.ChatPayload); ok {
		resp["Action"] = "pushchatmessage"
		resp["Address"] = chatMessage.Address
		resp["Username"] = chatMessage.UserName
		resp["Result"] = string(chatMessage.Content)
		ws.PushResult(resp)
	}
}
