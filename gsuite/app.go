package gsuite

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/zapr"
	"github.com/jlewi/gctl/config"
	"github.com/jlewi/monogo/gcp"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/gmail/v1"
)

// App is a struct that provides routines for configuring the application.
// It provides common application logic across subcommands.
type App struct {
	Config *config.Config
	Out    io.Writer
	TS     oauth2.TokenSource
}

func NewApp(out io.Writer) *App {
	return &App{
		Out: out,
	}
}

// LoadConfig loads the config. It takes an optional command. The command allows values to be overwritten from
// the CLI.
func (a *App) LoadConfig(cmd *cobra.Command) error {
	// N.B. at this point we haven't configured any logging so zap just returns the default logger.
	// TODO(jeremy): Should we just initialize the logger without cfg and then reinitialize it after we've read the config?
	if err := config.InitViper(cmd); err != nil {
		return err
	}
	cfg := config.GetConfig()

	if problems := cfg.IsValid(); len(problems) > 0 {
		fmt.Fprintf(os.Stdout, "Invalid configuration; %s\n", strings.Join(problems, "\n"))
		return fmt.Errorf("invalid configuration; fix the problems and then try again")
	}
	a.Config = cfg

	return nil
}

func (a *App) SetupLogging(logToFile bool) error {
	c := zap.NewDevelopmentConfig()

	lvl := a.Config.Logging.Level
	zapLvl := zap.NewAtomicLevel()
	err := zapLvl.UnmarshalText([]byte(lvl))
	if err != nil {
		return errors.Wrapf(err, "Could not convert level %v to ZapLevel", lvl)
	}

	// We write logs to a file in the log directory by default because we don't want to clutter the console.
	// We want to reserve the console for the output of the commands.
	logFile := filepath.Join(a.Config.GetLogDir(), "gctl.log")
	c.Level = zapLvl
	c.OutputPaths = []string{logFile}
	newLogger, err := c.Build()
	if err != nil {
		return errors.Wrapf(err, "Failed to initialize zap logger")
	}

	zap.ReplaceGlobals(newLogger)
	return nil
}

func (a *App) SetupTokenSource() error {
	if a.Config == nil {
		return errors.New("Config is nil; call LoadConfig first")
	}

	flow, err := gcp.NewWebFlowHelper(a.Config.OAuthClientFile, []string{gmail.GmailReadonlyScope, drive.DriveScope})
	if err != nil {
		return err
	}

	log := zapr.NewLogger(zap.L())

	cache := &gcp.FileTokenCache{
		CacheFile: a.Config.GetOAuthCredentialsFile(),
		Log:       log,
	}

	credsHelper := gcp.CachedCredentialHelper{
		CredentialHelper: flow,
		TokenCache:       cache,
		Log:              log,
	}
	ts, err := credsHelper.GetTokenSource(context.Background())
	if err != nil {
		return errors.Wrapf(err, "Error getting token source")
	}

	a.TS = ts
	return nil
}
