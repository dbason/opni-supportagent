package commands

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/dbason/opni-supportagent/pkg/cluster"
	"github.com/dbason/opni-supportagent/pkg/publish"
	"github.com/dbason/opni-supportagent/pkg/util"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

var (
	cleanupCluster bool
)

func BuildLocalCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "local",
		Short: "build local opnicluster in k3d and publish logs to it",
		RunE:  publishLocal,
	}
	command.Flags().BoolVar(&cleanupCluster, "cleanup", false, "delete the cluster after use")

	return command
}

func publishLocal(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("local requires exactly 1 argument; the distribution to use")
	}

	var publishFunc func(endpoint string) error
	switch Distribution(args[0]) {
	case RKE:
		publishFunc = publish.ShipRKEControlPlane
	case RKE2, K3S:
		return errors.New("distribution not currently supported")
	default:
		return errors.New("distribution must be one of rke, rke2, k3s")
	}

	err := cluster.CreateCluster(cmd.Context())
	if err != nil {
		return err
	}

	err = cluster.DeployCertManager(cmd.Context())
	if err != nil {
		return err
	}

	err = cluster.DeployOpniController(cmd.Context())
	if err != nil {
		return err
	}

	err = cluster.DeployOpni(cmd.Context())
	if err != nil {
		return err
	}

	portfowardCtx, cancel := context.WithCancel(cmd.Context())
	err = cluster.PayloadReceiverPort(portfowardCtx)
	if err != nil {
		cancel()
		return err
	}

	err = publishFunc(fmt.Sprintf("http://localhost:%d", cluster.PayloadReceiverForwardedPort))
	if err != nil {
		cancel()
		return err
	}
	cancel()

	portfowardCtx, cancel = context.WithCancel(cmd.Context())
	err = cluster.KibanaPort(portfowardCtx)
	if err != nil {
		cancel()
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	input := make(chan rune)
	intr := make(chan bool, 1)

	// Disable input buffer so we can read characters as they're entered
	setSttyState(bytes.NewBufferString("cbreak"))
	setSttyState(bytes.NewBufferString("-echo"))

	fmt.Printf("Kibana URL is http://localhost:%d\n", cluster.KibanaForwardedPort)
	fmt.Println("Press space to open the url in browser, or ctrl+c to exit")

	// signal goroutine
	go func() {
		s := <-sigs
		fmt.Println("Got signal", s)
		intr <- true
	}()
	// char input goroutine
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			c, _, err := reader.ReadRune()
			if err != nil {
				close(input)
				util.Log.Error(err)
			}
			input <- c
		}
	}()

	// label to handle goroutine signals
browser:
	for {
		select {
		case <-intr:
			break browser
		case c := <-input:
			if c == ' ' {
				browser.OpenURL(fmt.Sprintf("http://localhost:%d", cluster.KibanaForwardedPort))
			} else {
				fmt.Println("Unknown character entered")
			}
		}
	}

	// Set the terminal state back
	err = setSttyState(bytes.NewBufferString("-cbreak"))
	if err != nil {
		util.Log.Error(err)
	}
	err = setSttyState(bytes.NewBufferString("echo"))
	if err != nil {
		util.Log.Error(err)
	}

	// close port forward before deleting cluster
	cancel()
	if cleanupCluster {
		util.Log.Info("cleaning up cluster")
		cluster.DeleteCluster(cmd.Context())
	}

	return nil
}

func setSttyState(state *bytes.Buffer) (err error) {
	cmd := exec.Command("stty", state.String())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
