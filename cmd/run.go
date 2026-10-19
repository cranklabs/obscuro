package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/janklabs/obscuro/internal/crypto"
	"github.com/janklabs/obscuro/internal/store"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- COMMAND [args...]",
	Short: "Run a command with secrets loaded as environment variables",
	Long: `Decrypts all secrets in the selected namespace and adds them to the command's
environment. Secrets override existing environment variables with the same name.
Without --namespace/-n, only the default namespace is loaded.

The command runs directly, without a shell, in the current working directory.
Its stdin, stdout, stderr, and exit status are preserved. Your shell's environment
is not modified, and Obscuro does not write decrypted values to disk.

Place Obscuro flags before -- and the command and its arguments after it.`,
	Example: `  obscuro run -n my-app -- docker compose up -d
  obscuro run -- npm start
  obscuro run -n production -- sh -c './build.sh && ./deploy.sh'`,
	Args:                  cobra.MinimumNArgs(1),
	DisableFlagsInUseLine: true,
	SilenceUsage:          true,
	RunE: func(cmd *cobra.Command, args []string) error {
		key, err := authenticate()
		if err != nil {
			return err
		}
		secrets, err := store.LoadNamespaceSecrets(namespace)
		if err != nil {
			return err
		}

		env := os.Environ()
		for _, name := range store.ListKeys(secrets) {
			if name == "" || strings.ContainsAny(name, "=\x00") {
				return fmt.Errorf("secret %q cannot be used as an environment variable: name is empty or contains '=' or NUL", name)
			}
			value, err := crypto.Decrypt(key, secrets[name])
			if err != nil {
				return fmt.Errorf("decrypting secret %q: %w", name, err)
			}
			if strings.IndexByte(string(value), 0) >= 0 {
				return fmt.Errorf("secret %q cannot be used as an environment variable: value contains NUL", name)
			}
			env = append(env, name+"="+string(value))
		}

		child := exec.CommandContext(cmd.Context(), args[0], args[1:]...)
		// os/exec keeps the last value for duplicate keys (case-insensitively
		// on Windows), so namespace values override the inherited environment.
		child.Env = env
		child.Stdin = cmd.InOrStdin()
		child.Stdout = Stdout
		child.Stderr = cmd.ErrOrStderr()

		if err := runWithSignals(child); err != nil {
			var exit *exec.ExitError
			if errors.As(err, &exit) {
				code := exit.ExitCode()
				if status, ok := exit.Sys().(syscall.WaitStatus); ok && status.Signaled() {
					code = 128 + int(status.Signal())
				}
				return &exitErr{code: code, msg: fmt.Sprintf("command exited with status %d", code)}
			}
			return fmt.Errorf("running command: %w", err)
		}
		return nil
	},
}

// runWithSignals forwards interrupts and termination requests while waiting for
// the child. On Unix, this also handles signals sent only to Obscuro's PID.
func runWithSignals(child *exec.Cmd) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	if err := child.Start(); err != nil {
		return err
	}
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case sig := <-signals:
				_ = child.Process.Signal(sig)
			case <-done:
				return
			}
		}
	}()
	return child.Wait()
}

func init() {
	// Once the executable is encountered, remaining flags belong to it. The
	// explicit -- separator is recommended but not required for simple commands.
	runCmd.Flags().SetInterspersed(false)
	rootCmd.AddCommand(runCmd)
}
