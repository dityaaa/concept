package file

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepare(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "00001_create_user_table.adv.sql", "1 advance")
	writeFile(t, dir, "00002_create_user_table.rev.sql", "1 reverse")

}

func writeFile(t *testing.T, dir, name, content string) {
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
