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
	ctx := signal.SigTermCancelContext(context.Background())
	_, ctx, _, err := k8s.GetConfig(ctx, "auto", "")
	if err != nil {
		logrus.Infof("create k8s failed %s", err)
		return
	}
	
	os.Unsetenv("KUBECONFIG")
	
	<-ctx.Done()
	
	logrus.Info(ctx.Err())
	
	return
}
