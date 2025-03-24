package main

import (
	"flag"
	"fmt"
	"os"

	"path/filepath"
	"runtime"

	bizutils "{{RootModule}}/biz/utils"
	"{{RootModule}}/servers/{{MyServerName}}/internal/config"
	"{{RootModule}}/servers/{{MyServerName}}/internal/handler"
	"{{RootModule}}/servers/{{MyServerName}}/internal/middleware"
	"{{RootModule}}/servers/{{MyServerName}}/internal/svc"

	// "github.com/zeromicro/go-zero/core/conf"
	"github.com/fatih/color"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var (
	AppName   = "{{RootModule}}/{{MyServerName}}"
	Version   = "0.0.1.godev"
	GitCommit = "unknown:require build with `-ldflags \"-X main.BuildTime=$(git rev-parse --short HEAD)\""
	BuildTime = "unknown:require build with `-ldflags \"-X main.BuildTime=$(date +'%Y-%m-%dT%H:%M:%SZ')\""
)

var configFile = flag.String("f", "etc/{{MyServerName}}-api.yaml", "the config file")
var showVerbose = flag.Int("verbose", 0, "show build-in-flags and exit")
var initConfig = flag.String("init-config", "", "create config.yaml sample and exit")

func init() {
	//; runtime 运行时构建: 绑定到 flag
	//;; 运行时： pc, file, line, ok := runtime.Caller(0)
	//;; 提取目录部分
	flag.StringVar(&AppName, "show-appname", AppName, "程序名称")
	flag.StringVar(&Version, "show-version", Version, "程序版本")
	flag.StringVar(&BuildTime, "show-build-time", BuildTime, "构建时间戳")
	flag.StringVar(&GitCommit, "show-git-commit", GitCommit, "Git Commit Hash")
}

func main() {
	_, filename, _, _ := runtime.Caller(0)
	entryDir := filepath.Dir(filename)

	flag.Parse()

	if *showVerbose == 1 {
		fmt.Printf("%s\n", Version)
		os.Exit(0)
	}
	if *showVerbose >= 2 {
		fmt.Printf("应用名称: %s\n版本: %s\n提交: %s\n构建时间: %s\n",
			AppName, Version, GitCommit, BuildTime)
		os.Exit(0)
	}
	if *initConfig != "" {
		cfg_sample := filepath.Join(entryDir, *initConfig)
		if *initConfig == "." {
			cfg_sample = filepath.Join(entryDir, "etc/{{MyServerName}}-api-sample.yaml")
		}
		cfg_default := config.Config{}
		err2 := bizutils.RawYamlFile(cfg_sample, cfg_default)
		if err2 != nil {
			color.Red("ConfigSample: %s\n", cfg_sample)
			color.Red("InitConfig: Failed: %v\n", err2)
			os.Exit(1)
		} else {
			color.Green("InitConfig: Success: %s\n", cfg_sample)
			os.Exit(0)
		}

	}

	color.Green("[Start] %s: %s\n", AppName, Version)
	st_cfg, err := os.Stat(*configFile)
	if err != nil {
		color.Red("LoadConfigWarn: %v", err)
		color.Yellow("ExecDir: %s", entryDir)
		cfg2 := filepath.Join(entryDir, "etc/{{MyServerName}}-api.yaml")
		_, err = os.Stat(cfg2)
		if err == nil {
			color.Yellow("RollbackConfigOK: %s", cfg2)
			*configFile = cfg2
		} else {
			color.Red("RollbackConfigError: %s", &configFile)
			panic("ConfigFile Not Found!")
		}
	} else {
		color.Green("[Config]: %s, size: %v", *configFile, st_cfg.Size())
	}

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	//; 必须在服务初始化前调用
	//; 禁用 stat 日志
	if os.Getenv("LOG_ENABLE_STAT") == "0" {
		logx.DisableStat()
	}

	//; 初始化服务
	logx.MustSetup(c.LogConf)
	// fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	color.Yellow("GoServer(%s): Starting at < %s:%d > ...\n", c.Name, c.Host, c.Port)
	localip := bizutils.GetLocalIP()
	color.Green("服务启动(%s): http://%s:%d/api/info \n", c.Name, localip, c.Port)

	// config.LoadConfig(&c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	server.Use(middleware.HdlLogAccessMiddleware)

	server.Start()
}
