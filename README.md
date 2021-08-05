# cobra-update

Provides a [cobra](https://github.com/spf13/cobra) command to allow your apps to self-update themselves by downloading the latest GitHub release.

# Setup

1. Add the command to your Cobra app, passing in the owner/repo where new releases will be downloaded from:
    ```go
    import "github.com/Shamus03/cobra-update"

    func init() {
        rootCmd.AddCommand(update.Command("Shamus03", "munn"))
    }
    ```

2. (optional) Make sure your root `*cobra.Command` has its `Version` field set.  This is the most simple example:
    ```go
    var rootCmd = &cobra.Command {
        Use: "your-command",
        // other fields
        Version: "1.2.3",
    }
    ```
    - Skipping this step will cause the update command to always download the latest release, even if the executable is already up to date.
    - For a more complex example of setting the version through an automated workflow at build time, see [munn](https://github.com/Shamus03/munn)'s [version file](https://github.com/Shamus03/munn/blob/master/cmd/munn/version.go) and [semantic-release configuration](https://github.com/Shamus03/munn/blob/master/.releaserc.yml).