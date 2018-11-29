package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"syscall"
	"time"
	
    "k8s.io/client-go/tools/clientcmd"
	"github.com/docker/docker/pkg/reexec"
	"github.com/ehazlett/simplelog"
	"github.com/rancher/norman/pkg/dump"
	"github.com/rancher/norman/signal"
	"github.com/rancher/rancher/app"
	"github.com/rancher/rancher/pkg/logserver"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/golang/glog"
	"github.com/rancher/rancher/k8s"
)

var (
	VERSION = "dev"
)

func main() {
	app.RegisterPasswordResetCommand()
	app.RegisterEnsureDefaultAdminCommand()
	if reexec.Init() {
		return
	}

	os.Unsetenv("SSH_AUTH_SOCK")
	os.Unsetenv("SSH_AGENT_PID")
	os.Setenv("DISABLE_HTTP2", "true")

	if dir, err := os.Getwd(); err == nil {
		dmPath := filepath.Join(dir, "management-state", "bin")
		os.MkdirAll(dmPath, 0700)
		newPath := fmt.Sprintf("%s%s%s", dmPath, string(os.PathListSeparator), os.Getenv("PATH"))

		os.Setenv("PATH", newPath)
	}

	var config app.Config

	app := cli.NewApp()
	app.Version = VERSION
	app.Usage = "Complete container management platform"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "kubeconfig",
			Usage:       "Kube config for accessing k8s cluster",
			EnvVar:      "KUBECONFIG",
			Value:       "/var/lib/rancher/kube_config_cluster.yml",
			Destination: &config.KubeConfig,
		},
		cli.BoolFlag{
			Name:        "debug",
			Usage:       "Enable debug logs",
			Destination: &config.Debug,
		},
		cli.StringFlag{
			Name:        "add-local",
			Usage:       "Add local cluster (true, false, auto)",
			Value:       "auto",
			Destination: &config.AddLocal,
		},
		cli.IntFlag{
			Name:        "http-listen-port",
			Usage:       "HTTP listen port",
			Value:       8080,
			Destination: &config.HTTPListenPort,
		},
		cli.IntFlag{
			Name:        "https-listen-port",
			Usage:       "HTTPS listen port",
			Value:       8443,
			Destination: &config.HTTPSListenPort,
		},
		cli.StringFlag{
			Name:        "k8s-mode",
			Usage:       "Mode to run or access k8s API server for management API (embedded, external, auto)",
			Value:       "auto",
			Destination: &config.K8sMode,
		},
		cli.StringFlag{
			Name:  "log-format",
			Usage: "Log formatter used (json, text, simple)",
			Value: "text",
		},
		cli.StringSliceFlag{
			Name:  "acme-domain",
			Usage: "Domain to register with LetsEncrypt",
		},
		cli.BoolFlag{
			Name:  "no-cacerts",
			Usage: "Skip CA certs population in settings when set to true",
		},
		cli.StringFlag{
			Name:   "audit-log-path",
			EnvVar: "AUDIT_LOG_PATH",
			Value:  "/var/log/auditlog/rancher-api-audit.log",
			Usage:  "Log path for Rancher Server API. Default path is /var/log/auditlog/rancher-api-audit.log",
		},
		cli.IntFlag{
			Name:   "audit-log-maxage",
			Value:  10,
			EnvVar: "AUDIT_LOG_MAXAGE",
			Usage:  "Defined the maximum number of days to retain old audit log files",
		},
		cli.IntFlag{
			Name:   "audit-log-maxbackup",
			Value:  10,
			EnvVar: "AUDIT_LOG_MAXBACKUP",
			Usage:  "Defines the maximum number of audit log files to retain",
		},
		cli.IntFlag{
			Name:   "audit-log-maxsize",
			Value:  100,
			EnvVar: "AUDIT_LOG_MAXSIZE",
			Usage:  "Defines the maximum size in megabytes of the audit log file before it gets rotated, default size is 100M",
		},
		cli.IntFlag{
			Name:   "audit-level",
			Value:  0,
			EnvVar: "AUDIT_LEVEL",
			Usage:  "Audit log level: 0 - disable audit log, 1 - log event metadata, 2 - log event metadata and request body, 3 - log event metadata, request body and response body",
		},
	}

	app.Action = func(c *cli.Context) error {
		// enable profiler
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()

		config.ACMEDomains = c.GlobalStringSlice("acme-domain")
		config.NoCACerts = c.Bool("no-cacerts")

		config.AuditLevel = c.Int("audit-level")
		config.AuditLogPath = c.String("audit-log-path")
		config.AuditLogMaxage = c.Int("audit-log-maxage")
		config.AuditLogMaxbackup = c.Int("audit-log-maxbackup")
		config.AuditLogMaxsize = c.Int("audit-log-maxsize")
		initLogs(c, config)
		return run(config)
	}

	app.ExitErrHandler = func(c *cli.Context, err error) {
		logrus.Fatal(err)
	}

	app.Run(os.Args)
}

func initLogs(c *cli.Context, cfg app.Config) {
	if cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	switch c.String("log-format") {
	case "simple":
		logrus.SetFormatter(&simplelog.StandardFormatter{})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{})
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
	logrus.SetOutput(os.Stdout)
	logserver.StartServerWithDefaults()
}

func run(cfg app.Config) error {
	logrus.Infof("Rancher version %s is starting", VERSION)
	logrus.Infof("Rancher arguments %+v", cfg)
	glog.Infof("rancher is starting")
	dump.GoroutineDumpOn(syscall.SIGUSR1, syscall.SIGILL)
	ctx := signal.SigTermCancelContext(context.Background())
    
    //kubeConfig, err := clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
	_, ctx, kubeConfig, err := k8s.GetConfig(ctx, cfg.K8sMode, cfg.KubeConfig)
	if err != nil {
		return err
	}
	cfg.Embedded = true
	
	os.Unsetenv("KUBECONFIG")
	kubeConfig.Timeout = 30 * time.Second
	return app.Run(ctx, *kubeConfig, &cfg)
}
