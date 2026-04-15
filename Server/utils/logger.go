package utils

import (
	"cardGame/global"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/logger/interfaces"
	logruswrapper "github.com/topfreegames/pitaya/v2/logger/logrus"
)

var (
	logOnce       sync.Once
	log           *logrus.Logger
	normalLogHook *LevelHook
	pLogHook      *LevelHook
	sharedWriter  *RotateWriter
)

// 初始化pitaya用的log
func initPitayaLogger() {
	// set pitaya logger use json format logger
	logger.SetLogger(NewJsonLogger())
}

func initSharedWriter() error {
	maxAge, rTime, size := parserConfig()
	var err error
	sharedWriter, err = NewRotateWriter(
		global.GameConf.Logger.LogPath,  // 收集log的路徑
		global.GameConf.Game.HostName,   // host name 作為前綴
		time.Duration(maxAge)*time.Hour, // log 保留rotation 最長時間
		time.Duration(rTime)*time.Hour,  // rotation間隔
		size,                            // log 保留rotation 最大容量
	)
	return err
}

func NewJsonLogger() interfaces.Logger {

	var (
		pLog *logrus.Logger
		err  error
	)

	// 確保 shared writer 已初始化
	if sharedWriter == nil {
		if err := initSharedWriter(); err != nil {
			fmt.Printf("初始化共用 writer 失敗: %v\n", err)
			return nil
		}
	}

	pLog, pLogHook, err = SetupLogrusRotate(LogConfig{
		IsToStdout:   true,
		IsAsyncWrite: false,
		LogLevelSet:  getLogLevel(),
		FieldMap: map[string]string{
			"s_app": "pitaya",
			"s_pod": global.GameConf.Game.HostName,
		},
	}, true, sharedWriter)
	if err != nil {
		fmt.Printf("設置日誌系統失敗: %v\n", err)
	}

	return logruswrapper.NewWithFieldLogger(pLog)
}

// 要印log且本體的log初始化
func Log() *logrus.Logger {
	logOnce.Do(func() {

		var (
			game   = global.GameConf.Game
			stdOut = true
			err    error
		)

		// 是否要標準輸出
		if global.GameConf.Logger.LogStdOut != "" {
			stdOut, _ = strconv.ParseBool(global.GameConf.Logger.LogStdOut)
		}

		// 確保 shared writer 已初始化
		if sharedWriter == nil {
			if err := initSharedWriter(); err != nil {
				fmt.Printf("初始化共用 writer 失敗: %v\n", err)
				return
			}
		}

		log, normalLogHook, err = SetupLogrusRotate(LogConfig{
			IsToStdout:   stdOut,
			IsAsyncWrite: false,
			LogLevelSet:  getLogLevel(),
			FieldMap: map[string]string{
				"s_app": game.Connector,
				"s_pod": global.GameConf.Game.HostName,
			},
		}, false, sharedWriter)
		if err != nil {
			fmt.Printf("設置日誌系統失敗: %v\n", err)
		}
	})

	return log
}

// CloseLogHook 如果有啟用異步寫入log，則要等channel內的log寫完
func CloseLogHook() {
	normalLogHook.Wait()
	pLogHook.Wait()
}

// getLogLevel 取得設定檔的log層級
func getLogLevel() logrus.Level {
	var gameConf = global.GameConf

	level, err := logrus.ParseLevel(gameConf.Logger.Loglevel)
	if err != nil {
		panic(err)
	}
	return level
}

// parserConfig 根據設定黨或環境變數，取得log相關設定，若沒有則使用預設值
func parserConfig() (maxAge, rTime int, size int64) {

	maxAge, _ = strconv.Atoi(global.GameConf.Logger.LogMaxAge)
	if maxAge == 0 {
		maxAge = 24
	}
	rTime, _ = strconv.Atoi(global.GameConf.Logger.LogTime)
	if rTime == 0 {
		rTime = 1
	}
	size, _ = strconv.ParseInt(global.GameConf.Logger.LogSize, 10, 64)
	if size == 0 {
		size = 1 << 32
	}
	return
}
