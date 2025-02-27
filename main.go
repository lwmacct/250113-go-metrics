package main

import (
	"fmt"
	"os"

	"github.com/lwmacct/241224-go-template-pkgs/pkgs/m_cmd"
	"github.com/lwmacct/241224-go-template-pkgs/pkgs/m_log"
	"github.com/lwmacct/250113-go-metrics/app"
	"github.com/lwmacct/250113-go-metrics/app/start"
	"github.com/lwmacct/250113-go-metrics/app/test"
	"github.com/lwmacct/250113-go-metrics/app/version"
)

var mc *m_cmd.Ts

func main() {
	mc = m_cmd.New(nil)

	{
		// 命令行参数
		mc.AddCobra(version.Cmd().Cobra())
		mc.AddCobra(start.Cmd().Cobra())

		// 开发环境中的测试命令
		if os.Getenv("ACF_SHOW_TEST") == "1" {
			mc.AddCobra(test.Cmd().Cobra())
		}
	}

	{
		// 日志处理
		ml := m_log.NewConfig()
		ml.Level = app.Flag.Log.Level
		if app.Flag.Log.File == "" {
			app.Flag.Log.File = ml.File
		} else {
			ml.File = app.Flag.Log.File
		}
		if version.Workspace != "" {
			ml.CallerClip = version.Workspace
		}
		app.Log = m_log.NewTs(ml)
	}

	if err := mc.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
