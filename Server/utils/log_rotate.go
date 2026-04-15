package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/sirupsen/logrus"
)

func MappingLogrusLevel(lvl string) logrus.Level {

	// "panic" "fatal" "error" "warn" "info" "debug" "trace"
	switch lvl {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// RotateWriter 封裝 rotate writer 的相關設定
type RotateWriter struct {
	writer io.Writer
}

// NewRotateWriter 創建新的 RotateWriter
func NewRotateWriter(logPath, prefix string, maxAge, rotateTime time.Duration, rotateSize int64) (*RotateWriter, error) {
	err := ensureDir(logPath)
	if err != nil {
		return nil, fmt.Errorf("無法建立目錄 %s: %v", logPath, err)
	}

	var pre string
	if prefix != "" {
		pre = prefix + "-"
	}

	writer, err := rotatelogs.New(
		filepath.Join(logPath, pre+"%Y%m%d%H.json"),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(rotateTime),
		rotatelogs.WithRotationSize(rotateSize),
	)
	if err != nil {
		return nil, fmt.Errorf("無法設置日誌輪轉: %v", err)
	}

	return &RotateWriter{
		writer: writer,
	}, nil
}

// SetupLogrusRotate 設置 Logrus 與單一目錄的日誌輪轉
func SetupLogrusRotate(setting LogConfig, isPitaya bool, sharedWriter *RotateWriter) (*logrus.Logger, *LevelHook, error) {
	logIns := logrus.New()
	// 設定要紀錄的log level 以上
	logIns.SetLevel(setting.LogLevelSet)

	var formatter logrus.Formatter
	if isPitaya {
		// pitaya 用 formatter
		formatter = &PitayaFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
			FieldMap:        setting.FieldMap,
		}
	} else {
		// game server 用 formatter
		formatter = &GameFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z",
			FieldMap:        setting.FieldMap,
		}
	}
	logIns.SetFormatter(formatter)

	// 是否要os standard output
	if setting.IsToStdout {
		logIns.SetOutput(os.Stdout)
	} else {
		logIns.SetOutput(io.Discard)
	}

	// 會處理的log (全部)，會不會記錄看level setting
	levels := []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	}

	levelHook := NewLevelHook(formatter, levels, setting.IsAsyncWrite)

	for _, level := range levels {
		levelHook.AddWriter(level, sharedWriter.writer)
	}

	logIns.AddHook(levelHook)
	return logIns, levelHook, nil
}

type LogConfig struct {
	IsToStdout    bool
	IsAsyncWrite  bool
	LogNamePrefix string
	LogPath       string
	LogLevelSet   logrus.Level
	RotateMaxAge  time.Duration
	RotateTime    time.Duration
	RotateSize    int64
	FieldMap      map[string]string
}

type logEntry struct {
	level  logrus.Level
	data   []byte
	writer io.Writer
}

// LevelHook 自訂的 Hook，用於根據日誌級別將日誌寫入不同的檔案
type LevelHook struct {
	Writers    map[logrus.Level]io.Writer
	LevelSlice []logrus.Level
	Formatter  logrus.Formatter
	AsyncWrite bool
	logChan    chan logEntry
	wg         sync.WaitGroup
}

// GameFormatter provides custom level strings
type GameFormatter struct {
	TimestampFormat string
	FieldMap        map[string]string
}

// Format implements the logrus.Formatter interface
func (f *GameFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 定義自定義的日誌等級對應
	levelMap := map[logrus.Level]string{
		logrus.PanicLevel: "Fatal",
		logrus.FatalLevel: "Fatal",
		logrus.ErrorLevel: "Error",
		logrus.WarnLevel:  "Warning",
		logrus.InfoLevel:  "Information",
		logrus.DebugLevel: "Debug",
		logrus.TraceLevel: "Verbose",
	}

	// 創建輸出數據結構
	data := make(map[string]interface{})

	// 添加基本欄位
	data["s_i"] = getMsgId()
	data["s_t"] = entry.Time.UTC().Format(f.TimestampFormat)
	data["s_m"] = entry.Message
	data["s_l"] = levelMap[entry.Level]
	data["s_type"] = "Log"

	// 客製化的field欄位
	for k, v := range f.FieldMap {
		data[k] = v
	}

	// 序列化為 JSON
	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return append(serialized, '\n'), nil
}

// PitayaFormatter provides custom level strings
type PitayaFormatter struct {
	TimestampFormat string
	FieldMap        map[string]string
}

// Format implements the logrus.Formatter interface
func (f *PitayaFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 定義自定義的日誌等級對應
	levelMap := map[logrus.Level]string{
		logrus.PanicLevel: "Fatal",
		logrus.FatalLevel: "Fatal",
		logrus.ErrorLevel: "Error",
		logrus.WarnLevel:  "Warning",
		logrus.InfoLevel:  "Information",
		logrus.DebugLevel: "Debug",
		logrus.TraceLevel: "Verbose",
	}

	// 創建輸出數據結構
	data := make(map[string]interface{})

	// 添加基本欄位
	data["s_i"] = getMsgId()
	data["s_t"] = entry.Time.UTC().Format(f.TimestampFormat)
	data["s_l"] = levelMap[entry.Level]
	data["s_type"] = "Log"

	// 客製化的field欄位
	for k, v := range f.FieldMap {
		data[k] = v
	}

	// pitaya 內建的log紀錄會放到withField
	// 抽出來丟進message
	for k, v := range entry.Data {

		if slices.Contains([]string{"route", "requestId", "userId"}, k) {
			entry.Message += fmt.Sprintf(", %s : %v", k, v)
		}
	}

	data["s_m"] = entry.Message

	// 序列化為 JSON
	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return append(serialized, '\n'), nil
}

// NewLevelHook 建立新的 LevelHook
func NewLevelHook(formatter logrus.Formatter, levels []logrus.Level, asyncWrite bool) *LevelHook {
	hook := &LevelHook{
		Writers:    make(map[logrus.Level]io.Writer),
		LevelSlice: levels,
		Formatter:  formatter,
		AsyncWrite: asyncWrite,
		logChan:    make(chan logEntry, 1000), // 使用帶緩衝的 channel
	}

	if asyncWrite {
		hook.startWorker()
	}

	return hook
}

// startWorker 啟動一個 worker goroutine 來處理異步寫入
func (hook *LevelHook) startWorker() {
	hook.wg.Add(1)
	go func() {
		defer hook.wg.Done()
		for entry := range hook.logChan {
			_, err := entry.writer.Write(entry.data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing log: %v\n", err)
			}
		}
	}()
}

// Fire 實現 Fire 方法，當有日誌輸出時會被呼叫
func (hook *LevelHook) Fire(entry *logrus.Entry) error {
	writer, ok := hook.Writers[entry.Level]
	if !ok {
		return fmt.Errorf("未找到级别 %v 的 writer", entry.Level)
	}

	b, err := hook.Formatter.Format(entry)
	if err != nil {
		return err
	}

	if hook.AsyncWrite {
		select {
		case hook.logChan <- logEntry{level: entry.Level, data: b, writer: writer}:
			// 成功發送到 channel
		default:
			// channel 已滿，改為同步寫入
			_, err = writer.Write(b)
			if err != nil {
				return err
			}
		}
	} else {
		_, err = writer.Write(b)
		if err != nil {
			return err
		}
	}

	return nil
}

// Levels 實現 Levels 方法，返回 Hook 會處理的日誌級別
func (hook *LevelHook) Levels() []logrus.Level {
	return hook.LevelSlice
}

// AddWriter 添加 writer 到 Hook 中
func (hook *LevelHook) AddWriter(level logrus.Level, writer io.Writer) {
	hook.Writers[level] = writer
}

// Wait 等待所有異步寫入完成
func (hook *LevelHook) Wait() {
	if hook.AsyncWrite {
		close(hook.logChan)
		hook.wg.Wait()
	}
}

// 確保目錄存在的輔助函式
func ensureDir(dirName string) error {
	err := os.MkdirAll(dirName, 0755)
	if err != nil {
		return err
	}
	return nil
}
