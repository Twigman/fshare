package store

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

type fakeMultipartFile struct {
	*bytes.Reader
}

func (f *fakeMultipartFile) Close() error { return nil }

func TestFileService_SaveUploadedFile(t *testing.T) {
	uploadDir := t.TempDir()

	// Inhalt der Testdatei
	content := []byte("Hallo Welt!")
	file := &fakeMultipartFile{bytes.NewReader(content)}

	// Test-Resource
	res := &Resource{
		Name:              "test.txt",
		IsPrivate:         true,
		OwnerHashedKey:    "owner123",
		AutoDeleteInHours: 0,
	}

	// Fake-DB einhängen
	fakeDB := &FakeDatabase{}
	fs := NewFileService(uploadDir, fakeDB)

	// Funktion ausführen
	uuid, err := fs.SaveUploadedFile(file, res)
	if err != nil {
		t.Fatalf("SaveUploadedFile fehlgeschlagen: %v", err)
	}

	// Datei existiert?
	savedPath := filepath.Join(uploadDir, "test.txt")
	data, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("Datei wurde nicht gespeichert: %v", err)
	}

	if !bytes.Equal(data, content) {
		t.Errorf("Inhalt falsch.\nGot:  %q\nWant: %q", data, content)
	}

	// DB-Aufruf prüfen
	if !fakeDB.Called {
		t.Error("FakeDB.SaveFile wurde nicht aufgerufen")
	}
	if fakeDB.LastName != "test.txt" {
		t.Errorf("Erwarteter Name %q, aber bekam %q", "test.txt", fakeDB.LastName)
	}
	if fakeDB.LastOwner != "owner123" {
		t.Errorf("Owner falsch: %s", fakeDB.LastOwner)
	}
	if fakeDB.LastUUID != uuid {
		t.Errorf("UUID falsch: %s ≠ %s", fakeDB.LastUUID, uuid)
	}
}
