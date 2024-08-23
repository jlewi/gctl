package gsuite

import (
	"context"
	"os"
	"testing"
)

func Test_Search(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping test in GitHub Actions")
	}

	app := NewApp(os.Stdout)
	if err := app.LoadConfig(nil); err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if err := app.SetupTokenSource(); err != nil {
		t.Fatalf("Error setting up token source: %v", err)
	}

	inbox, err := NewInbox(*app.Config, app.TS)
	if err != nil {
		t.Fatalf("Error creating inbox: %v", err)
	}

	results, err := inbox.Search(context.Background(), "from:me", 10)
	if err != nil {
		t.Fatalf("Error searching inbox: %v", err)
	}
	t.Logf("Found %d messages", len(results.Messages))
}
