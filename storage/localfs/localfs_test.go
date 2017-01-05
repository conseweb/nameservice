package localfs

import (
	"testing"
)

type testData struct {
	val   string
	isErr bool
}

func TestAbx(t *testing.T) {
	chroot := "/tmp/ss"
	fsdriver, err := NewDriver(chroot)
	if err != nil {
		t.Fatal(err)
		return
	}

	fs, ok := fsdriver.(*Driver)
	if !ok {
		t.Fatal("can not transfer Driver.")
		return
	}

	for k, v := range map[string]testData{
		"":         testData{chroot, false},
		"/a":       testData{chroot + "/a", false},
		"/a/../b":  testData{chroot + "/b", false},
		"/..":      testData{chroot + "", true},
		"/a/..":    testData{chroot + "", false},
		"/a/../..": testData{chroot + "", true},
	} {
		val, err := fs.Abs(k)
		if v.isErr {
			if err == nil {
				t.Errorf("%s should be returned error", k)
			}
		} else if val != v.val {
			t.Errorf("%s should be %s, but %s", k, v.val, val)
		}
	}
}
