package snaptel

import (
	"fmt"
	"os"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	snapclient "github.com/intelsdi-x/snap-client-go/client"
	"github.com/intelsdi-x/snap-client-go/client/operations"
	"github.com/urfave/cli"
)

const (
	errNoClient = "Error: no Client is created"
)

var opClient *operations.Client

type UsageError struct {
	s   string
	ctx *cli.Context
}

func (ue UsageError) Error() string {
	return ue.s
}

func (ue UsageError) Help() {
	cli.ShowCommandHelp(ue.ctx, ue.ctx.Command.Name)
}

func newUsageError(s string, ctx *cli.Context) UsageError {
	return UsageError{s, ctx}
}

func GetOperationClient(host, basePath, scheme string) *operations.Client {
	transport := httptransport.New(host, basePath, []string{scheme})
	tc := snapclient.New(transport, strfmt.Default)

	return tc.Operations
}

func getHost() string {
	host := os.Getenv("SNAP_CLIENT_GO_HOST") + ":8181"
	if host == ":8181" {
		host = "127.0.0.1:8181"
	}
	return host
}

func SetOperationsClient(c *operations.Client) {
	opClient = c
}

func getOperationsClient() *operations.Client {
	return opClient
}

// GetFirstChar gets the first character of a giving string.
func GetFirstChar(s string) string {
	firstChar := ""
	for _, r := range s {
		firstChar = fmt.Sprintf("%c", r)
		break
	}
	return firstChar
}
