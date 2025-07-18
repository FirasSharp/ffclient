// MIT License
//
// Copyright (c) 2025 Firas Jelassi
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package main provides the command-line interface for the ffclient application,
// which is a client for downloading multiple files from https://fuckingfast.co/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/FirasSharp/ffclient/cmd"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/spf13/cobra"
)

// Options contains the configuration options for the download client
type Options struct {
	SavePath  string // Destination directory for downloaded files
	InputFile string // Text file containing URLs to download (one per line)
	Links     string // Comma-separated URLs to download
}

var opts Options

// Version information that will be set during build
var (
	version = "0.1.2-dev" // default version for development builds
	commit  = "none"      // git commit hash
	date    = "unknown"   // build date
)

// rootCmd defines the main command for the ffclient application
var rootCmd = &cobra.Command{
	Use:     "ff",
	Short:   "Client for multi download from https://fuckingfast.co/",
	Long:    `Client made to download multiple files from the https://fuckingfast.co/ hosting service.`,
	Version: fmt.Sprintf("%s\ncommit: %s\nbuilt: %s", version, getCommit(), getDate()),
	Run: func(_ *cobra.Command, _ []string) {
		if err := cmd.ExecuteMultiDownload(opts.SavePath, opts.InputFile, opts.Links); err != nil {
			log.Entry().WithError(err).Fatal()
		}
		log.Entry().Info("Successfully executed!")
	},
}

// main is the entry point for the application
func main() {
	defaultPath, err := getDefaultDownloadPath()
	if err != nil {
		log.Entry().Errorf("Error getting download path: %v\n", err)
		return
	}
	rootCmd.Flags().StringVarP(&opts.SavePath, "savePath", "s", defaultPath, "Destination directory for downloaded files.")
	rootCmd.Flags().StringVarP(&opts.InputFile, "inputFile", "i", "", "Text file containing URLs to download (one fuckingfast.co URL per line)")
	rootCmd.Flags().StringVarP(&opts.Links, "links", "l", "", "Comma-separated fuckingfast.co URLs (e.g., \"https://fuckingfast.co/file1,https://fuckingfast.co/file2\")")
	rootCmd.MarkFlagsMutuallyExclusive("inputFile", "links")
	rootCmd.MarkFlagsOneRequired("inputFile", "links")

	if err := rootCmd.Execute(); err != nil {
		log.Entry().Error(err)
		os.Exit(1)
	}
}

// getCommit returns the git commit hash for the current build.
// It first checks the build-time set commit variable, and if not available,
// falls back to reading build info to get the VCS revision.
// Returns the short (7-character) commit hash if available.
func getCommit() string {
	if commit != "" && commit != "none" {
		return commit
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:7] // return short commit hash
			}
		}
	}
	return commit
}

// getDate returns the build date in a human-readable format.
// It first checks the build-time set date variable, and if not available,
// falls back to reading build info to get the VCS time.
// Returns the current time if no build date information is available.
func getDate() string {
	if date != "" && date != "unknown" {
		return date
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.time" {
				if t, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					return t.Format("2006-01-02 15:04:05")
				}
			}
		}
	}
	return time.Now().Format("2006-01-02 15:04:05")
}

// getDefaultDownloadPath determines the appropriate default download directory
// based on the operating system. For Windows it uses USERPROFILE\Downloads,
// for Linux it first checks XDG_DOWNLOAD_DIR, and for all other systems
// it falls back to ~/Downloads.
// Returns an error if the home directory cannot be determined.
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
