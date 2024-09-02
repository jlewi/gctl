package gsuite

import (
	"context"
	"os"
	"testing"
)

func Test_ImportHTML(t *testing.T) {
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

	d, err := NewDrive(*app.Config, app.TS)
	if err != nil {
		t.Fatalf("Error creating drive: %v", err)
	}

	if err := d.ImportHTMLToGoogleDoc(context.Background(), "/tmp/notes.html", "notes.html"); err != nil {
		t.Fatalf("Error importing the html document: %v", err)
	}
}
