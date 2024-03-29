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
	"github.com/kerraform/kegistry/internal/driver/local"
	"github.com/kerraform/kegistry/internal/driver/s3"
	"github.com/kerraform/kegistry/internal/logging"
	"github.com/kerraform/kegistry/internal/metric"
	"github.com/kerraform/kegistry/internal/server"
	"github.com/kerraform/kegistry/internal/trace"
	v1 "github.com/kerraform/kegistry/internal/v1"
	"github.com/kerraform/kegistry/internal/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	otracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
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

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceVersionKey.String(version.Version),
			semconv.ServiceNameKey.String(cfg.Name),
		),
	)
	if err != nil {
		logger.Error("failed to setup the otel resource", zap.Error(err))
		return err
	}

	var tp *otracesdk.TracerProvider
	if cfg.Trace.Enable {
		var sexp otracesdk.SpanExporter
		switch trace.ExporterType(cfg.Trace.Type) {
		case trace.ExporterTypeConsole:
			sexp, err = trace.NewConsoleExporter(os.Stdout)
		case trace.ExporterTypeJaeger:
			sexp, err = trace.NewJaegerExporter(cfg.Trace.Jaeger.Endpoint)
		default:
			return fmt.Errorf("trace type %s not supported", cfg.Trace.Type)
		}
		if err != nil {
			logger.Error("failed to setup the trace provider", zap.Error(err))
			return err
		}

		logger.Info("setup otel tracer", zap.String("trace", cfg.Trace.Type))
		tp = trace.NewTracer(r, sexp)
		otel.SetTracerProvider(tp)
	} else {
		logger.Debug("tracing disabled")
		tp = trace.NewTracer(r, nil)
	}
	t := tp.Tracer(cfg.Trace.Name)

	logger.Info("setup backend", zap.String("backend", cfg.Backend.Type), zap.String("rootPath", cfg.Backend.RootPath))
	var d *driver.Driver
	switch driver.DriverType(cfg.Backend.Type) {
	case driver.DriverTypeS3:
		d, err = s3.NewDriver(logger, &s3.DriverOpts{
			AccessKey:    cfg.Backend.S3.AccessKey,
			Bucket:       cfg.Backend.S3.Bucket,
			Endpoint:     cfg.Backend.S3.Endpoint,
			SecretKey:    cfg.Backend.S3.SecretKey,
			Tracer:       t,
			UsePathStyle: cfg.Backend.S3.UsePathStyle,
		})

		if err != nil {
			return err
		}
	case driver.DriverTypeLocal:
		d = local.NewDriver(&local.DriverConfig{
			Logger:   logger,
			Tracer:   t,
			RootPath: cfg.Backend.RootPath,
		})
	default:
		return fmt.Errorf("backend type %s not supported", cfg.Backend.Type)
	}

	metrics := metric.New(logger, d)

	wg, ctx := errgroup.WithContext(ctx)

	v1 := v1.New(&v1.HandlerConfig{
		Driver: d,
		Logger: logger,
	})

	svr := server.NewServer(&server.ServerConfig{
		Driver:         d,
		EnableModule:   cfg.EnableModule,
		EnableProvider: cfg.EnableProvider,
		Logger:         logger,
		Metric:         metrics,
		Tracer:         t,
		V1:             v1,
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

	// Context for shutdown
	newCtx := context.Background()
	if err := svr.Shutdown(newCtx); err != nil {
		logger.Error("failed to graceful shutdown server", zap.Error(err))
		return err
	}

	if tp != nil {
		if err := tp.Shutdown(newCtx); err != nil {
			logger.Error("failed to shutdown trace provider", zap.Error(err))
			return err
		}
	}

	return nil
}
