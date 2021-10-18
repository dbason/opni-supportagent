package commands

import (
	"bufio"
	"bytes"
	"context"
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

func BuildLocalCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "local",
		Short: "build local opnicluster in k3d and publish logs to it",
		RunE:  publishLocal,
	}

	return command
}

func publishLocal(cmd *cobra.Command, args []string) error {
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

	err = publish.ShipRKEControlPlane(fmt.Sprintf("http://localhost:%d", cluster.PayloadReceiverForwardedPort))
	if err != nil {
		cancel()
		return err
	}
	cancel()

	portfowardCtx, cancel = context.WithCancel(cmd.Context())
	defer cancel()
	err = cluster.KibanaPort(portfowardCtx)
	if err != nil {
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
				fmt.Println("\nUnknown character entered")
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

	return nil
}

func setSttyState(state *bytes.Buffer) (err error) {
	cmd := exec.Command("stty", state.String())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
