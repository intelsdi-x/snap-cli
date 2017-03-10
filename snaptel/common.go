package snaptel

import (
	"fmt"

	"github.com/intelsdi-x/snap-client-go/snap"
	"github.com/urfave/cli"
)

const (
	errNoClient = "Error: no Client is created"
)

var snapClient *snap.Client

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

// SetClient provides a way to set the private snapClient in this package.
func SetClient(c *snap.Client) {
	snapClient = c
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
