package proto

import (
	"github.com/spf13/cobra"

	"github.com/cnsync/kratos/cmd/kratos/internal/proto/add"
	"github.com/cnsync/kratos/cmd/kratos/internal/proto/client"
	"github.com/cnsync/kratos/cmd/kratos/internal/proto/server"
)

// CmdProto represents the proto command.
var CmdProto = &cobra.Command{
	Use:   "proto",
	Short: "Generate the proto files",
	Long:  "Generate the proto files.",
}

func init() {
	CmdProto.AddCommand(add.CmdAdd)
	CmdProto.AddCommand(client.CmdClient)
	CmdProto.AddCommand(server.CmdServer)
}
