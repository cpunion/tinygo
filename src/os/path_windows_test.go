// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os_test

import (
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestFixLongPath(t *testing.T) {
	if os.CanUseLongPaths {
		return
	}
	// 248 is long enough to trigger the longer-than-248 checks in
	// fixLongPath, but short enough not to make a path component
	// longer than 255, which is illegal on Windows. (which
	// doesn't really matter anyway, since this is purely a string
	// function we're testing, and it's not actually being used to
	// do a system call)
	veryLong := "l" + strings.Repeat("o", 248) + "ng"
	for _, test := range []struct{ in, want string }{
		// Short; unchanged:
		{`C:\short.txt`, `C:\short.txt`},
		{`C:\`, `C:\`},
		{`C:`, `C:`},
		// The "long" substring is replaced by a looooooong
		// string which triggers the rewriting. Except in the
		// cases below where it doesn't.
		{`C:\long\foo.txt`, `\\?\C:\long\foo.txt`},
		{`C:/long/foo.txt`, `\\?\C:\long\foo.txt`},
		{`C:\long\foo\\bar\.\baz\\`, `\\?\C:\long\foo\bar\baz`},
		{`\\unc\path`, `\\unc\path`},
		{`long.txt`, `long.txt`},
		{`C:long.txt`, `C:long.txt`},
		{`c:\long\..\bar\baz`, `c:\long\..\bar\baz`},
		{`\\?\c:\long\foo.txt`, `\\?\c:\long\foo.txt`},
		{`\\?\c:\long/foo.txt`, `\\?\c:\long/foo.txt`},
	} {
		in := strings.ReplaceAll(test.in, "long", veryLong)
		want := strings.ReplaceAll(test.want, "long", veryLong)
		if got := os.FixLongPath(in); got != want {
			got = strings.ReplaceAll(got, veryLong, "long")
			t.Errorf("fixLongPath(%q) = %q; want %q", test.in, got, test.want)
		}
	}
}

// TODO: bring back upstream version's TestMkdirAllLongPath once os.RemoveAll and t.TempDir implemented

// isWine returns true if executing on wine (Wine Is Not an Emulator), which
// is compatible with windows but does not reproduce all its quirks.
func isWine() bool {
	return os.Getenv("WINEUSERNAME") != ""
}

func TestMkdirAllExtendedLength(t *testing.T) {
	// TODO: revert to upstream version once os.RemoveAll and t.TempDir implemented
	tmpDir := os.TempDir()

	const prefix = `\\?\`
	if len(tmpDir) < 4 || tmpDir[:4] != prefix {
		fullPath, err := syscall.FullPath(tmpDir)
		if err != nil {
			t.Fatalf("FullPath(%q) fails: %v", tmpDir, err)
		}
		tmpDir = prefix + fullPath
	}
	path := tmpDir + `\dir\`
	if err := os.MkdirAll(path, 0777); err != nil {
		t.Fatalf("MkdirAll(%q) failed: %v", path, err)
	}

	if isWine() {
		// TODO: use t.Skip once implemented
		t.Log("wine: Skipping check for no-dots-for-you quirk in windows extended paths")
		return
	}
	path = path + `.\dir2`
	if err := os.MkdirAll(path, 0777); err == nil {
		t.Fatalf("MkdirAll(%q) should have failed, but did not", path)
	}
}
