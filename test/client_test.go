// +build e2e

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kelseyhightower/envconfig"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

const (
	vcsim     = "vcsim"
	job       = "client"
	secret    = "vsphere-credentials"
	mountPath = "/var/bindings/vsphere"
)

func TestWaitForClientJob(t *testing.T) {
	clientFeature := features.New("appsv1/deployment").WithLabel("env", "e2e").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			depl, svc := newSimulator(cfg.Namespace())
			secr := newVCSecret(cfg.Namespace(), secret, "user", "pass") // vcsim defaults

			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			if err = client.Resources().Create(ctx, depl); err != nil {
				t.Fatal(err)
			}

			if err = client.Resources().Create(ctx, svc); err != nil {
				t.Fatal(err)
			}

			if err = client.Resources().Create(ctx, secr); err != nil {
				t.Fatal(err)
			}

			// wait for the deployment to become ready
			err = wait.For(conditions.New(client.Resources()).ResourceMatch(depl, func(object k8s.Object) bool {
				d := object.(*appsv1.Deployment)
				return d.Status.AvailableReplicas == 1
			}), wait.WithTimeout(time.Minute*3))

			if err != nil {
				t.Fatal(err)
			}

			t.Log("vcsim ready", "replicas", "1")

			return ctx
		}).
		Assess("job completes", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			c := newClient(cfg.Namespace(), secret)

			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			if err = client.Resources().Create(ctx, c); err != nil {
				t.Fatal(err)
			}

			// wait for the c to succeed
			if err = wait.For(conditions.New(client.Resources()).JobCompleted(c), wait.WithTimeout(time.Minute*3)); err != nil {
				t.Fatal(err)
			}

			t.Log("client job complete")

			return ctx
		}).
		Feature()

	testenv.Test(t, clientFeature)
}

func newSimulator(namespace string) (*appsv1.Deployment, *v1.Service) {
	l := map[string]string{
		"app":  vcsim,
		"test": "e2e",
	}

	sim := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vcsim,
			Namespace: namespace,
			Labels:    l,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: l,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: l,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  vcsim,
						Image: "vmware/vcsim:latest",
						Args: []string{
							"/vcsim",
							"-l",
							":8989",
						},
						ImagePullPolicy: v1.PullIfNotPresent,
						Ports: []v1.ContainerPort{
							{
								Name:          "https",
								ContainerPort: 8989,
							},
						},
					}},
				},
			},
		},
	}

	svc := v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vcsim,
			Namespace: namespace,
			Labels:    l,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "https",
					Port: 443,
					TargetPort: intstr.IntOrString{
						IntVal: 8989,
					},
				},
			},
			Selector: l,
		},
	}

	return &sim, &svc
}

func newVCSecret(namespace, name, username, password string) *v1.Secret {
	l := map[string]string{
		"test": "e2e",
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    l,
		},
		Data: map[string][]byte{
			v1.BasicAuthUsernameKey: []byte(username),
			v1.BasicAuthPasswordKey: []byte(password),
		},
		Type: v1.SecretTypeBasicAuth,
	}
}

func newClient(namespace, secret string) *batchv1.Job {
	var e envConfig

	if err := envconfig.Process("", &e); err != nil {
		panic("process environment variables: " + err.Error())
	}

	l := map[string]string{
		"app":  job,
		"test": "e2e",
	}

	k8senv := []v1.EnvVar{
		{Name: "VCENTER_URL", Value: fmt.Sprintf("https://%s.%s.svc.cluster.local", vcsim, namespace)},
		{Name: "VCENTER_INSECURE", Value: "true"},
	}

	client := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      job,
			Namespace: namespace,
			Labels:    l,
		},
		Spec: batchv1.JobSpec{
			Parallelism: pointer.Int32(1),
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: l,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:            job,
						Image:           fmt.Sprintf("%s/client", e.DockerRepo),
						Env:             k8senv,
						ImagePullPolicy: v1.PullIfNotPresent,
						VolumeMounts: []v1.VolumeMount{
							{
								Name:      "credentials",
								ReadOnly:  true,
								MountPath: mountPath,
							},
						},
					}},
					Volumes: []v1.Volume{{
						Name: "credentials",
						VolumeSource: v1.VolumeSource{
							Secret: &v1.SecretVolumeSource{
								SecretName: secret,
							},
						},
					}},
					RestartPolicy:                 v1.RestartPolicyOnFailure,
					TerminationGracePeriodSeconds: pointer.Int64Ptr(5),
				},
			},
		},
	}

	return &client
}
