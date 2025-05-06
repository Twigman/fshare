package fake

import "bytes"

type FakeMultipartFile struct {
	*bytes.Reader
}

func (f *FakeMultipartFile) Close() error { return nil }
