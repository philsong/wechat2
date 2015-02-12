// @description wechat2 是腾讯微信公众平台 api 的 golang 语言封装
// @link        https://github.com/philsong/wechat2 for the canonical source repository
// @license     https://github.com/philsong/wechat2/blob/master/LICENSE
// @authors     chanxuehong(chanxuehong@gmail.com)

package jssdk

import (
	"errors"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/philsong/wechat2/mp"
)

const defaultTickDuration = time.Minute // 获取 jsapi_ticket 失败时尝试获取的间隔时间.

var _ TicketServer = new(DefaultTicketServer)

// TicketServer 的简单实现.
//  NOTE:
//  一般用于单进程环境, 因为 DefaultTicketServer 同时也实现了一个简单的中控服务器, 而不是简单的
//  实现了 TicketServer 接口, 所以整个系统只能存在一个 DefaultTicketServer 实例!!!
type DefaultTicketServer struct {
	mp.WechatClient

	// 缓存最后一次从微信服务器获取的 jsapi_ticket 的结果.
	// goroutine ticketAutoUpdate() 里有个定时器, 每次触发都会更新 currentTicket.
	currentTicket struct {
		rwmutex sync.RWMutex
		ticket  string
		err     error
	}

	// goroutine ticketAutoUpdate() 监听 resetTicketRefreshTickChan,
	// 如果有新的数据, 则重置定时器, 定时时间为 resetTicketRefreshTickChan 传过来的数据.
	resetTicketRefreshTickChan chan time.Duration
}

// 创建一个新的 DefaultTicketServer.
//  如果 httpClient == nil 则默认使用 http.DefaultClient.
func NewDefaultTicketServer(tokenServer mp.TokenServer, httpClient *http.Client) (srv *DefaultTicketServer) {
	if tokenServer == nil {
		panic("nil tokenServer")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	srv = &DefaultTicketServer{
		WechatClient: mp.WechatClient{
			TokenServer: tokenServer,
			HttpClient:  httpClient,
		},
		resetTicketRefreshTickChan: make(chan time.Duration),
	}

	// 获取 jsapi_ticket 并启动 goroutine ticketAutoUpdate
	resp, err := srv.getTicket()
	if err != nil {
		srv.currentTicket.ticket = ""
		srv.currentTicket.err = err
		go srv.ticketAutoUpdate(defaultTickDuration)
	} else {
		srv.currentTicket.ticket = resp.Ticket
		srv.currentTicket.err = nil
		go srv.ticketAutoUpdate(time.Duration(resp.ExpiresIn) * time.Second)
	}

	return
}

func (srv *DefaultTicketServer) Ticket() (ticket string, err error) {
	srv.currentTicket.rwmutex.RLock()
	ticket = srv.currentTicket.ticket
	err = srv.currentTicket.err
	srv.currentTicket.rwmutex.RUnlock()
	return
}

func (srv *DefaultTicketServer) TicketRefresh() (ticket string, err error) {
	resp, err := srv.getTicket()
	if err != nil {
		srv.currentTicket.rwmutex.Lock()
		srv.currentTicket.ticket = ""
		srv.currentTicket.err = err
		srv.currentTicket.rwmutex.Unlock()

		srv.resetTicketRefreshTickChan <- defaultTickDuration
		return
	} else {
		srv.currentTicket.rwmutex.Lock()
		srv.currentTicket.ticket = resp.Ticket
		srv.currentTicket.err = nil
		srv.currentTicket.rwmutex.Unlock()

		ticket = resp.Ticket
		srv.resetTicketRefreshTickChan <- time.Duration(resp.ExpiresIn) * time.Second
		return
	}
}

// 单独一个 goroutine 来定时获取 jsapi_ticket.
//  tickDuration: 启动后初始 tickDuration.
func (srv *DefaultTicketServer) ticketAutoUpdate(tickDuration time.Duration) {
	var ticker *time.Ticker

NEW_TICK_DURATION:
	ticker = time.NewTicker(tickDuration)
	for {
		select {
		case tickDuration = <-srv.resetTicketRefreshTickChan:
			ticker.Stop()
			goto NEW_TICK_DURATION

		case <-ticker.C:
			resp, err := srv.getTicket()
			if err != nil {
				srv.currentTicket.rwmutex.Lock()
				srv.currentTicket.ticket = ""
				srv.currentTicket.err = err
				srv.currentTicket.rwmutex.Unlock()

				// 出错则重置到 defaultTickDuration
				if tickDuration != defaultTickDuration {
					ticker.Stop()
					tickDuration = defaultTickDuration
					goto NEW_TICK_DURATION
				}
			} else {
				srv.currentTicket.rwmutex.Lock()
				srv.currentTicket.ticket = resp.Ticket
				srv.currentTicket.err = nil
				srv.currentTicket.rwmutex.Unlock()

				newTickDuration := time.Duration(resp.ExpiresIn) * time.Second
				if tickDuration != newTickDuration {
					ticker.Stop()
					tickDuration = newTickDuration
					goto NEW_TICK_DURATION
				}
			}
		}
	}
}

type ticketResponse struct {
	Ticket    string `json:"ticket"`     // 获取到的 jsapi_ticket
	ExpiresIn int64  `json:"expires_in"` // jsapi_ticket 的有效时间，单位：秒
}

// 从微信服务器获取 jsapi_ticket.
func (srv *DefaultTicketServer) getTicket() (resp *ticketResponse, err error) {
	var result struct {
		mp.Error
		ticketResponse
	}

	incompleteURL := "https://api.weixin.qq.com/cgi-bin/ticket/getticket?type=jsapi&access_token="
	if err = srv.GetJSON(incompleteURL, &result); err != nil {
		return
	}

	if result.ErrCode != mp.ErrCodeOK {
		err = &result.Error
		return
	}

	// 由于网络的延时, jsapi_ticket 过期时间留了一个缓冲区;
	// 正常情况下微信服务器会返回 7200, 则缓冲区的大小为 10 分钟.
	switch {
	case result.ExpiresIn > 60*60:
		result.ExpiresIn -= 60 * 10
		resp = &result.ticketResponse
		return
	case result.ExpiresIn > 60*30:
		result.ExpiresIn -= 60 * 5
		resp = &result.ticketResponse
		return
	case result.ExpiresIn > 60*5:
		result.ExpiresIn -= 60
		resp = &result.ticketResponse
		return
	case result.ExpiresIn > 60:
		result.ExpiresIn -= 10
		resp = &result.ticketResponse
		return
	case result.ExpiresIn > 0:
		resp = &result.ticketResponse
		return
	default:
		err = errors.New("invalid expires_in: " + strconv.FormatInt(result.ExpiresIn, 10))
		return
	}
}
