package main

import (
	"os"

	v1 "micro-scheduler/api/helloworld/v1"
	"micro-scheduler/internal/biz"
	"micro-scheduler/internal/conf"
	"micro-scheduler/internal/data"
	"micro-scheduler/internal/service"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hashicorp/consul/api"
)

func main() {
	logger := log.NewStdLogger(os.Stdout)
	log := log.NewHelper(logger)

	// 加载配置文件
	c := config.New(
		config.WithSource(
			file.NewSource("../../configs/config.yaml"),
		),
	)
	if err := c.Load(); err != nil {
		log.Fatal(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		log.Fatal(err)
	}

	// 初始化数据层和服务层
	dataData, cleanup, err := data.NewData(bc.Data, logger) // 使用配置文件中的数据配置
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	greeterRepo := data.NewGreeterRepo(dataData, logger)
	greeterUsecase := biz.NewGreeterUsecase(greeterRepo, logger)
	greeterService := service.NewGreeterService(greeterUsecase)

	// consul client
	config := api.DefaultConfig()
	config.Address = "120.27.227.150:8500"
	consulClient, err := api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	r := consul.New(consulClient) // 把 consulClient 客户端连接添加到 go-kratos 中的registry

	// http server
	httpSrv := http.NewServer(
		http.Address(bc.Server.Http.Addr), // 使用配置文件中的HTTP端口
		http.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	// grpc server
	grpcSrv := grpc.NewServer(
		grpc.Address(bc.Server.Grpc.Addr), // 使用配置文件中的gRPC端口
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)

	v1.RegisterGreeterServer(grpcSrv, greeterService)     // grpc 方式调用方法
	v1.RegisterGreeterHTTPServer(httpSrv, greeterService) // http 方式调用方法

	app := kratos.New(
		kratos.Name("micro-scheduler"),
		kratos.Server(
			grpcSrv,
			httpSrv,
		),
		kratos.Registrar(r), // 这里用 consul 作为服务发现中心
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
