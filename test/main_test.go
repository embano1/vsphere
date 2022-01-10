// +build e2e

package e2e

import (
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var testenv env.Environment

type envConfig struct {
	KindCluster string `envconfig:"KIND_CLUSTER_NAME" required:"true" default:"e2e"`
	DockerRepo  string `envconfig:"KO_DOCKER_REPO" required:"true" default:"kind.local"`
}

func TestMain(m *testing.M) {
	var e envConfig

	if err := envconfig.Process("", &e); err != nil {
		panic("process environment variables: " + err.Error())
	}

	testenv = env.New()
	namespace := envconf.RandomName("kind-ns", 16)
	testenv.Setup(
		envfuncs.CreateKindCluster(e.KindCluster),
		envfuncs.CreateNamespace(namespace),
	)
	testenv.Finish(
		envfuncs.DeleteNamespace(namespace),
		// envfuncs.DestroyKindCluster(e.KindCluster),
	)
	os.Exit(testenv.Run(m))
}
