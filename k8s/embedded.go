// +build k8s

package k8s

import (
	"context"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rancher/pkg/embedded"
	"github.com/rancher/rancher/pkg/kubectl"
	"k8s.io/client-go/rest"
	"fmt"
)

func init() {
	reexec.Register("/usr/bin/kubectl", kubectl.Main)
	fmt.Print("reexec kubectl")
	reexec.Register("kubectl", kubectl.Main)
	reexec.Register("kubelet", kubectl.Main)
	reexec.Register("kube-proxy", kubectl.Main)
}

func getEmbedded(ctx context.Context) (bool, context.Context, *rest.Config, error) {
	ctx, kubeConfig, err := embedded.Run(ctx)
	if err != nil {
		return true, ctx, nil, err
	}

	restConfig, err := getExternal(kubeConfig)
	return true, ctx, restConfig, err
}
