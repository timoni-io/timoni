package permissions

type RepoPerm uint8

// Git Repo permissions:
//
//go:generate stringer -type=RepoPerm
const (
	Repo_View           RepoPerm = iota // mozliwosc podglądu danego git-repo w trybie RO
	Repo_Pull                           // mozliwosc pobrania
	Repo_Push                           // mozliwosc pushowania
	Repo_RemoteManage                   // mozliwosc zarządzania zdalnymi repozytoriami
	Repo_LocalManage                    // mozliwosc zarządzania lokalnymi repozytoriami
	Repo_SettingsManage                 // mozliwosc zmiany ustawień danego git-repo
	__Repo_Iter                         // used for iteration, do not remove
)
