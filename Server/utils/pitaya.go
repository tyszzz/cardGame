package utils

import (
	"cardGame/global"
	"context"
	"errors"
	"sync"

	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/acceptorwrapper"
	"github.com/topfreegames/pitaya/v2/config"
	"github.com/topfreegames/pitaya/v2/interfaces"
	"github.com/topfreegames/pitaya/v2/session"
)

var (
	onceAction sync.Once
	app        pitaya.Pitaya
	conn       string
	sp         session.SessionPool
)

func App() pitaya.Pitaya {
	onceAction.Do(func() {
		var (
			vp              = viper.GetViper()
			origConf        = config.NewConfig(vp)
			conn            = global.GameConf.Game.Connector
			builder         = pitaya.NewBuilderWithConfigs(true, conn, pitaya.Standalone, map[string]string{}, origConf)
			rateLimitConfig = config.NewPitayaConfig(origConf).Conn.RateLimiting
			origAcceptor    = acceptor.NewWSAcceptor(viper.GetString("fk.port"))
			acceptorLimit   = acceptorwrapper.WithWrappers(
				origAcceptor,
				acceptorwrapper.NewRateLimitingWrapper(builder.MetricsReporters, rateLimitConfig))
		)
		builder.AddAcceptor(acceptorLimit)
		app = builder.Build()
		sp = builder.SessionPool
	})
	return app
}

func GroupBroadcast(group string, remoteFunc string, data interface{}) {
	err := App().GroupBroadcast(context.Background(), conn, group, remoteFunc, data)
	if err != nil {
		GroupBroadcastError(group, err)
	}
}

// 有發生錯誤檢查是誰有錯，然後發送測試
func GroupBroadcastError(group string, err error) {
	// pitaya已經有印errorlog了，server就不印
	groupMember, _ := App().GroupMembers(context.Background(), group)
	for _, memberUid := range groupMember {
		// 嘗試push一個message過去
		if err := SessionPush(memberUid, "test", ""); err != nil {
			// 有發生送不到直接移除這個member
			GroupRemoveMember(group, memberUid)
			Log().Warnf("[GroupBroadcastError] group:%s got error from broadcast, remove memberUid:%s", group, memberUid)
		}
	}

}

func GroupRemoveMember(group string, uid string) {
	err := App().GroupRemoveMember(context.Background(), group, uid)
	if err != nil {
		Log().Errorf("%s, group remove member err round id:%s, err:%v", group, uid, err)
	}
}

func GroupAddMember(group string, uid string) {
	App().GroupAddMember(context.Background(), group, uid)
}

func GroupCountMembers(group string) (int, error) {
	return App().GroupCountMembers(context.Background(), group)
}

func GroupCreate(group string) {
	App().GroupCreate(context.Background(), group)
}

func GetModule(m string) interfaces.Module {
	mod, err := App().GetModule(m)
	if err != nil {
		panic(err)
	}
	return mod
}

func GetSessionFromCtx(ctx context.Context) session.Session {
	return App().GetSessionFromCtx(ctx)
}

func GetSessionByUID(sid string) session.Session {
	return sp.GetSessionByUID(sid)
}

func SessionPush(sid string, route string, v interface{}) error {
	//s := sp.GetSessionByUID(sid)
	//if s != nil {
	//	return s.Push(route, v)
	//}
	//return errors.New("session not found")
	_, err := App().SendPushToUsers(route, v, []string{sid}, conn)
	return err
}

func SessionClose(sid string) error {
	s := sp.GetSessionByUID(sid)
	if s != nil {
		s.Close()
		return nil
	}
	return errors.New("session not found")
}
