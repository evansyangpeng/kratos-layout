package main

import (
	"flag"
	"os"

	pb "github.com/go-kratos/kratos-layout/api/helloworld/v1"
	"github.com/go-kratos/kratos-layout/internal/service"
	"github.com/go-kratos/kratos/v2"
	grpcconf "github.com/go-kratos/kratos/v2/api/kratos/config/grpc"
	httpconf "github.com/go-kratos/kratos/v2/api/kratos/config/http"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/source/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/log/stdlog"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string
	// Branch is current branch name the code is built off.
	Branch string
	// Revision is the short commit hash of source tree.
	Revision string
	// BuildDate is the date when the binary was built.
	BuildDate string
)

var flagconf string

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func main() {
	flag.Parse()
	conf := config.New(config.WithSource(
		file.NewSource(flagconf),
	))
	if err := conf.Load(); err != nil {
		panic(err)
	}

	logger := stdlog.NewLogger(stdlog.Writer(os.Stdout))
	defer logger.Close()

	log := log.NewHelper("main", logger)
	log.Infow(
		"version", Version,
		"branch", Branch,
		"revision", Revision,
		"build_date", BuildDate,
	)

	// build transport server
	hc := new(httpconf.Server)
	gc := new(grpcconf.Server)
	if err := conf.Value("http.server").Scan(hc); err != nil {
		panic(err)
	}
	if err := conf.Value("grpc.server").Scan(gc); err != nil {
		panic(err)
	}
	httpSrv := http.NewServer(http.Apply(hc))
	grpcSrv := grpc.NewServer(grpc.Apply(gc))

	// register service
	gs := service.NewGreeterService()
	pb.RegisterGreeterServer(grpcSrv, gs)
	pb.RegisterGreeterHTTPServer(httpSrv, gs)

	// application lifecycle
	app := kratos.New()
	app.Append(httpSrv)
	app.Append(grpcSrv)

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		log.Errorf("start failed: %v\n", err)
	}
}
