package utils

type Locker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type FakeLock struct{}

func (l *FakeLock) Lock()    {}
func (l *FakeLock) Unlock()  {}
func (l *FakeLock) RLock()   {}
func (l *FakeLock) RUnlock() {}
