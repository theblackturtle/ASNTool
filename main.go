package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ASNTool",
	Short: "Simple ASN Lookup Tool",
}

var asnCmd = &cobra.Command{
	Use:     "asn",
	Long:    "Get ASN informations from IP Addresses",
	Example: "ASNTool asn [IP] [IP] ...",
	Run:     getAsn,
}

var netCmd = &cobra.Command{
	Use:     "net",
	Long:    "Get IP ranges from ASNs",
	Example: "ASNTool net [ASN] [ASN] ... (without AS prefix)",
	Run:     getNetBlocks,
}

func main() {
	rootCmd.AddCommand(asnCmd)
	rootCmd.AddCommand(netCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getAsn(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
	for _, arg := range args {
		address := net.ParseIP(arg)
		if address == nil {
			continue
		}
		asnRecord, err := IPToASRecord(address.String())
		if err != nil {
			printError(fmt.Sprintf("[-] IP Address: %s\n", address.String()))
			continue
		}
		printSuccess(fmt.Sprintf("[+] IP Address: %s\n", address.String()))
		fmt.Println("ASN:", asnRecord.ASN)
		fmt.Println("Prefix:", asnRecord.Prefix)
		fmt.Println("ASName:", asnRecord.ASName)
		fmt.Println("CN:", asnRecord.CN)
		fmt.Println("ISP:", asnRecord.ISP)
		fmt.Println()
	}
}

func getNetBlocks(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		os.Exit(0)
	}
	for _, arg := range args {
		asn, err := strconv.Atoi(arg)
		if err != nil {
			continue
		}
		asnNetBlocks, err := ASNToNetblocks(asn)
		if err != nil {
			printError(fmt.Sprintf("[-] AS: %d\n", asn))
			continue
		}
		printSuccess(fmt.Sprintf("[+] AS: %d\n", asn))
		for _, nb := range asnNetBlocks {
			fmt.Println(nb)
		}
		fmt.Println()
	}
}

func printError(msg string) {
	color.New(color.FgHiRed).Fprintf(os.Stderr, msg)
}

func printSuccess(msg string) {
	color.New(color.FgHiGreen).Fprintf(os.Stderr, msg)
}
