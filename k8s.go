package main

import (
	"context"
	"os"
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
		return err
	}
	
	os.Unsetenv("KUBECONFIG")
	
	<-ctx.Done()
	
	return ctx.Err()
}
