package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/support/utils"
)

func TestE2E(t *testing.T) {
	fleetlockDeploymentFeat := features.New("deploy fleetlock").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			err = decoder.ApplyWithManifestDir(ctx, r, "manifests/release", "deployment.yaml", []resources.CreateOption{})
			if err != nil {
				t.Fatalf("Failed to apply fleetlock manifest: %v", err)
			}

			return ctx
		}).
		Assess("available", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var dep appsv1.Deployment
			if err := c.Client().Resources().Get(ctx, "fleetlock", namespace, &dep); err != nil {
				t.Fatalf("Failed to get fleetlock-deployment: %v", err)
			}

			err := wait.For(conditions.New(c.Client().Resources()).DeploymentConditionMatch(&dep, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
			if err != nil {
				t.Fatalf("Failed to wait for fleetlock deployment to be created: %v", err)
			}

			pods := &corev1.PodList{}
			err = c.Client().Resources(namespace).List(ctx, pods)
			if err != nil || pods.Items == nil {
				t.Fatalf("Error while getting pods: %v", err)
			}
			if len(pods.Items) != 2 {
				t.Fatalf("Not enough fleetlock pods, expected 2 but got %d", len(pods.Items))
			}
			for _, pod := range pods.Items {
				err = wait.For(conditions.New(c.Client().Resources()).PodConditionMatch(&pod, corev1.PodReady, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
				if err != nil {
					t.Fatalf("Failed to wait for pod %s to be ready: %v", pod.GetName(), err)
				}
			}

			return ctx
		}).
		Feature()

	var url string

	nodePortFeat := features.New("add node port for testing").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			service := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nodeport",
					Namespace: namespace,
					Labels: map[string]string{
						"app": "fleetlock",
					},
				},
				Spec: corev1.ServiceSpec{
					Selector: map[string]string{
						"app": "fleetlock",
					},
					Type: corev1.ServiceTypeNodePort,
					Ports: []corev1.ServicePort{
						{
							Name:     "http",
							Protocol: corev1.ProtocolTCP,
							Port:     8080,
							NodePort: nodePort,
						},
					},
				},
			}

			err := c.Client().Resources().Create(ctx, service)
			if err != nil {
				t.Fatalf("Failed to create node port: %v", err)
			}

			return ctx
		}).
		Assess("ip", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var node corev1.Node
			err := c.Client().Resources().Get(ctx, clusterName+"-control-plane", "", &node)
			if err != nil {
				t.Fatalf("Failed to get node %s: %v", clusterName+"-control-plane", err)
			}

			var nodeIP string
			for _, addr := range node.Status.Addresses {
				if addr.Type == corev1.NodeInternalIP {
					nodeIP = addr.Address
					break
				}
			}

			if nodeIP == "" {
				t.Fatalf("Failed to find nodes internal ip")
			}
			url = "http://" + nodeIP + ":" + strconv.Itoa(nodePort)

			return ctx
		}).
		Assess("available", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			assert.Eventually(t, func() bool {
				res, err := http.Get(url + "/healthz")
				if err != nil {
					fmt.Printf("Fleetlock is not yet reachable on %s: %v", url, err)
					return false
				}
				return res.StatusCode == http.StatusOK
			}, time.Minute*5, time.Second*30, "Fleetlock should be available via node port")

			return ctx
		}).
		Feature()

	testenv.Test(t, fleetlockDeploymentFeat, nodePortFeat)
	if t.Failed() {
		t.Fatal("Failed to deploy fleetlock, can't run tests")
	}

	t.Run("fleetctl", func(t *testing.T) {
		assert := assert.New(t)

		exitCode, err := execFleetctl("lock", url)
		assert.Equal(0, exitCode, "Should lock the slot")
		assert.NoError(err, "Should lock the slot")

		exitCode, err = execFleetctl("lock", url)
		assert.Equal(0, exitCode, "Should show the slot as locked")
		assert.NoError(err, "Should show the slot as locked")

		exitCode, err = execFleetctl("lock -g master -i abcdef", url)
		assert.Equal(0, exitCode, "Should lock the slot in master group")
		assert.NoError(err, "Should lock the slot in master group")

		exitCode, err = execFleetctl("lock -i abcdef", url)
		assert.Equal(1, exitCode, "Should fail to lock the slot, as it is already locked")
		assert.Error(err, "Should fail to lock the slot, as it is already locked")

		exitCode, err = execFleetctl("release -g master -i abcdef", url)
		assert.Equal(0, exitCode, "Should release the slot in master group")
		assert.NoError(err, "Should release the slot in master group")

		exitCode, err = execFleetctl("release", url)
		assert.Equal(0, exitCode, "Should release the slot in default group")
		assert.NoError(err, "Should release the slot in default group")

		exitCode, err = execFleetctl("lock -i abcdef", url)
		assert.Equal(0, exitCode, "Should lock the slot")
		assert.NoError(err, "Should lock the slot")
	})
}

func execFleetctl(args, url string) (int, error) {
	err := utils.RunCommandWithSeperatedOutput("bin/"+fleetctlBinary+" "+args+" "+url, os.Stdout, os.Stderr)

	exitCode := 0
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			exitCode = exiterr.ExitCode()
		}
	}
	return exitCode, err
}
