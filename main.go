package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/FirasSharp/ffclient/cmd"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

type Options struct {
	savePath  string
	inputFile string
	links     string
}

var opts Options

var rootCmd = &cobra.Command{
	Use:   "ff",
	Short: "Client for multi download from https://fuckingfast.co/",
	Long:  `Client made to download multiple files from the  https://fuckingfast.co/ hosting service.`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := cmd.ExecuteMultiDownload(opts.savePath, opts.inputFile, opts.links); err != nil {
			log.Entry().WithError(err).Fatal()
		}
		log.Entry().Info("Successfully executed!")
	},
}

func main() {
	defaultPath, err := getDefaultDownloadPath()
	if err != nil {
		log.Entry().Errorf("Error getting download path: %v\n", err)
		return
	}
	rootCmd.Flags().StringVar(&opts.savePath, "savePath", defaultPath, "Destination directory for downloaded files.")
	rootCmd.Flags().StringVar(&opts.inputFile, "inputFile", "", "Text file containing URLs to download (one fuckingfast.co URL per line)")
	rootCmd.Flags().StringVar(&opts.links, "links", "", "Comma-separated fuckingfast.co URLs (e.g., \"https://fuckingfast.co/file1,https://fuckingfast.co/file2\")")
	rootCmd.MarkFlagsMutuallyExclusive("inputFile", "links")
	rootCmd.MarkFlagsOneRequired("inputFile", "links")
	if err := rootCmd.Execute(); err != nil {
		log.Entry().Error(err)
		os.Exit(1)
	}
}

func getDefaultDownloadPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		dir := os.Getenv("USERPROFILE")
		if dir == "" {
			return "", fmt.Errorf("USERPROFILE environment variable not set")
		}
		return filepath.Join(dir, "Downloads"), nil
	case "linux":
		dir := os.Getenv("XDG_DOWNLOAD_DIR")
		if dir != "" {
			return dir, nil
		}
	}
	// Fallback for mac, linux if XDG_DOWNLOAD_DIR is not set and other unix-like OSes
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Downloads"), nil
}
