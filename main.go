package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/kerraform/kegistry/internal/config"
	"github.com/kerraform/kegistry/internal/driver"
	"github.com/kerraform/kegistry/internal/logging"
	"github.com/kerraform/kegistry/internal/server"
	v1 "github.com/kerraform/kegistry/internal/v1"
	"github.com/kerraform/kegistry/internal/version"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	exitOk = iota
	exitError
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}

	os.Exit(exitOk)
}

func run(args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load(ctx)
	if err != nil {
		return err
	}

	logger, err := logging.NewLogger(os.Stdout, logging.Level(cfg.Log.Level), logging.Format(cfg.Log.Format))
	if err != nil {
		return err
	}

	logger = logger.With(
		zap.String("version", version.Version),
		zap.String("revision", version.Commit),
	)

	logger.Info("setup backend", zap.String("backend", cfg.Backend.Type))

	var opts []driver.DriverOpt

	switch driver.DriverType(cfg.Backend.Type) {
	case driver.DriverTypeS3:
		opts = append(opts, driver.WithS3(&driver.S3Opts{
			AccessKey: cfg.Backend.S3.AccessKey,
			Bucket:    cfg.Backend.S3.Bucket,
			Endpoint:  cfg.Backend.S3.Endpoint,
			SecretKey: cfg.Backend.S3.SecretKey,
		}))
	}

	driver, err := driver.NewDriver(driver.DriverType(cfg.Backend.Type), logger.Named("driver"), opts...)
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)

	v1 := v1.New(&v1.HandlerConfig{
		Driver: driver,
		Logger: logger,
	})

	svr := server.NewServer(&server.ServerConfig{
		Driver: driver,
		Logger: logger,
		V1:     v1,
	})

	conn, err := net.Listen("tcp", cfg.Address())
	if err != nil {
		return err
	}

	logger.Info("server started", zap.Int("port", cfg.Port))
	wg.Go(func() error {
		return svr.Serve(ctx, conn)
	})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)
	select {
	case v := <-sigCh:
		logger.Info("received signal %d", zap.String("signal", v.String()))
	case <-ctx.Done():
	}

	if err := svr.Shutdown(ctx); err != nil {
		logger.Error("failed to graceful shutdown server", zap.Error(err))
		return err
	}

	return nil
}
