package permissions

type GlobPerm uint8

// Global permissions:
//
//go:generate stringer -type=GlobPerm
const (
	Glob_AccessToWebUI           GlobPerm = iota // mozliwosc zalogowania sie do panelu webUI Timoni, brak tego = ban
	Glob_AccessToAdminZone                       // mozliwosc dostępu do strey admina, która jest tylko RO
	Glob_AccessToKube                            // bezposredni dostep do kubka - ten sam kubeconfig co Timoni
	Glob_ManageGlobalMemebers                    // mozliwosc zarzadzania globalnymi zespołami i uczestnikami zespołów
	Glob_CreateAndDeleteEnvs                     // mozliwosc tworzenia i usuwania srodowisk
	Glob_CreateAndDeleteGitRepos                 // mozliwosc tworzenia i usuwania git-reposow
	__Glob_Iter                                  // used for iteration, do not remove
)
