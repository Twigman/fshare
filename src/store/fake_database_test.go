package store

type FakeDatabase struct {
	Called      bool
	LastName    string
	LastUUID    string
	LastOwner   string
	LastPrivate bool
}

func (f *FakeDatabase) init() error {
	return nil
}

func (f *FakeDatabase) saveResource(r Resource) error {
	return nil
}

func (f *FakeDatabase) saveFile(uuid, name string, isPrivate bool, owner string, autoDel int) error {
	f.Called = true
	f.LastName = name
	f.LastUUID = uuid
	f.LastPrivate = isPrivate
	f.LastOwner = owner
	return nil
}

func (f *FakeDatabase) saveAPIKey(hash string, comment string) error {
	return nil
}
