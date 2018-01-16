package cmd

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/FreifunkBremen/yanic/respond"
	"github.com/FreifunkBremen/yanic/runtime"
	"github.com/spf13/cobra"
)

var (
	wait      int
	port      int
	ipAddress string
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:     "query <interfaces> <destination>",
	Short:   "Sends a query on the interface to the destination and waits for a response",
	Example: `yanic query "eth0,wlan0" "fe80::eade:27ff:dead:beef"`,
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ifaces := strings.Split(args[0], ",")
		dstAddress := net.ParseIP(args[1])

		log.Printf("Sending request address=%s ifaces=%s", dstAddress, ifaces)

		var ifacesConfigs []respond.InterfaceConfig
		for _, iface := range ifaces {
			ifaceConfig := respond.InterfaceConfig{
				InterfaceName: iface,
				Port:          port,
				IPAddress:     ipAddress,
			}
			ifacesConfigs = append(ifacesConfigs, ifaceConfig)
		}

		nodes := runtime.NewNodes(&runtime.NodesConfig{})

		collector := respond.NewCollector(nil, nodes, []string{}, ifacesConfigs)
		defer collector.Close()
		collector.SendPacket(dstAddress)

		time.Sleep(time.Second * time.Duration(wait))

		for id, data := range nodes.List {
			log.Printf("%s: %+v", id, data)
		}
	},
}

func init() {
	RootCmd.AddCommand(queryCmd)
	queryCmd.Flags().IntVar(&wait, "wait", 1, "Seconds to wait for a response")
	queryCmd.Flags().IntVar(&port, "port", 0, "define a port to listen (if not set or set to 0 the kernel will use a random free port at its own)")
	queryCmd.Flags().StringVar(&ipAddress, "ip", "", "ip address which is used for sending (optional - without definition used the link-local address)")
}
