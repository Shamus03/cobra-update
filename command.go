package update

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func Command(owner, repo string, options ...Option) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Download the latest release from GitHub",
		Long: `Download the latest release from GitHub and install it in-place.

If the GITHUB_TOKEN environment variable is set, it will be used for any GitHub API requests.

GITHUB_TOKEN is required for private repositories.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := update(cmd, owner, repo, options); err != nil {
				fmt.Fprintf(cmd.OutOrStderr(), "Error: %v", err)
			}
		},
	}
	cmd.Flags().Bool("debug", false, "show debug logs")

	return cmd
}

func update(cmd *cobra.Command, owner, repo string, options []Option) (updateErr error) {
	ctx := context.Background()

	debugOut := io.Discard
	if debug, _ := cmd.Flags().GetBool("debug"); debug {
		debugOut = cmd.OutOrStdout()
	}

	oc := &optionCtx{
		logger:      log.New(cmd.OutOrStdout(), "", 0),
		debugLogger: log.New(debugOut, "DEBUG: ", log.Lmsgprefix),
		errorLogger: log.New(cmd.OutOrStderr(), "ERROR: ", log.Lmsgprefix),
		assetIsCompatibleFunc: func(asset *github.ReleaseAsset) bool {
			return strings.Contains(asset.GetName(), runtime.GOOS)
		},
	}
	for _, o := range options {
		o.apply(oc)
	}

	logger := oc.logger
	debugLogger := oc.debugLogger
	errorLogger := oc.errorLogger
	isCompatible := oc.assetIsCompatibleFunc

	var tc *http.Client
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc = oauth2.NewClient(ctx, ts)
	}
	client := github.NewClient(tc)

	release, resp, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		if resp.StatusCode == 404 {
			return fmt.Errorf("getting latest release (if this is a private repo, you need to set the GITHUB_TOKEN environment variable): %v", err)
		}
		return fmt.Errorf("getting latest release: %v", err)
	}
	currentVersion := cmd.Root().Version

	debugLogger.Printf("Found release %s, current version is %s", release.GetName(), currentVersion)

	if release.GetName() == currentVersion {
		logger.Printf("Already up to date\n")
		return nil
	}

	logger.Printf("Updating to %s\n", release.GetName())

	var assetID int64
	for _, asset := range release.Assets {
		debugLogger.Printf("Asset name: %s, id: %d, download at: %s", asset.GetName(), asset.GetID(), asset.GetBrowserDownloadURL())
		if isCompatible(asset) {
			debugLogger.Printf("Will download asset %d", asset.GetID())
			assetID = asset.GetID()
			break
		}
	}

	if assetID == 0 {
		return fmt.Errorf("could not find a suitable release to download, use --debug flag for details")
	}

	assetReader, _, err := client.Repositories.DownloadReleaseAsset(ctx, owner, repo, assetID, http.DefaultClient)
	if err != nil {
		return fmt.Errorf("download release asset: %v", err)
	}
	defer assetReader.Close()

	outPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %v", err)
	}
	debugLogger.Printf("Got asset response, will write to %s", outPath)

	tmpDir, err := os.MkdirTemp("", repo+"-bak-")
	if err != nil {
		return fmt.Errorf("make backup dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	bakFile := filepath.Join(tmpDir, filepath.Base(outPath))
	debugLogger.Printf("Creating backup file: %s", bakFile)

	if err := os.Rename(outPath, bakFile); err != nil {
		return fmt.Errorf("rename old executable: %v", err)
	}
	defer func() {
		if updateErr != nil {
			if err := os.Rename(outPath, bakFile); err != nil {
				errorLogger.Printf("failed to restore backup file after installation error: %v", err)
			}
		}
	}()

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create new executable: %v", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, assetReader); err != nil {
		return fmt.Errorf("write new executable: %v", err)
	}

	return nil
}
