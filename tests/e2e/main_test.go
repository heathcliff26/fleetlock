package e2e

import (
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
	"sigs.k8s.io/e2e-framework/support/utils"
)

const namespace = "fleetlock"

const nodePort = 32080

var (
	clusterName    string
	fleetctlBinary string

	testenv env.Environment
)

func TestMain(m *testing.M) {
	testenv = env.New()
	clusterName = envconf.RandomName("fleetlock-e2e", 24)
	fleetctlBinary = envconf.RandomName("fleetctl-e2e", 24)

	err := os.Chdir("../..")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = utils.RunCommandWithSeperatedOutput("make REPOSITORY=localhost TAG="+clusterName+" image", os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = utils.RunCommandWithSeperatedOutput("make REPOSITORY=localhost TAG="+clusterName+" manifests", os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = utils.RunCommandWithSeperatedOutput("hack/build.sh fleetctl "+fleetctlBinary, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	imageArchive := fmt.Sprintf("tmp_fleetlock_image_%s.tar", clusterName)

	err = utils.RunCommandWithSeperatedOutput(fmt.Sprintf("podman save -o %s localhost/fleetlock:%s", imageArchive, clusterName), os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), clusterName),
		envfuncs.LoadImageArchiveToCluster(clusterName, imageArchive),
	)
	testenv.Finish(
		envfuncs.ExportClusterLogs(clusterName, "./logs"),
		envfuncs.DestroyCluster(clusterName),
	)

	exitCode := testenv.Run(m)
	if exitCode != 0 {
		fmt.Printf("Failed e2e testsuite with exit code %d\n", exitCode)
	}

	fmt.Print("\nRunning cleanup\n\n")

	fmt.Printf("Removing image archive file %s\n", imageArchive)
	err = os.Remove(imageArchive)
	if err != nil {
		fmt.Printf("Failed to remove image archive %s: %v\n", imageArchive, err)
		os.Exit(1)
	}

	fmt.Printf("Removing fleetctl binary %s\n", "bin/"+fleetctlBinary)
	err = os.Remove("bin/" + fleetctlBinary)
	if err != nil {
		fmt.Printf("Failed to remove fleetctl binary %s: %v\n", "bin/"+fleetctlBinary, err)
		os.Exit(1)
	}

	fmt.Printf("Deleting kind cluster %s\n", clusterName)
	err = utils.RunCommandWithSeperatedOutput("kind delete cluster --name "+clusterName, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("")

	os.Exit(exitCode)
}
