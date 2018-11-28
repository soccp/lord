package main

import (
	"context"
	"os"
	"github.com/sirupsen/logrus"
	"github.com/rancher/norman/signal"
	"github.com/rancher/rancher/k8s"

)

var (
	VERSION = "dev"
)

func main() {
	logrus.Info("k8s is starting")
	ctx := signal.SigTermCancelContext(context.Background())
	embedded, ctx, kubeConfig, err := k8s.GetConfig(ctx, "auto", "")
	if err != nil {
		logrus.Infof("create k8s failed %s", err)
		return
	}
	logrus.Info(embedded)
	logrus.Info(kubeconfig)
	os.Unsetenv("KUBECONFIG")
	
	<-ctx.Done()
	
	logrus.Info(ctx.Err())
	
	return
}
