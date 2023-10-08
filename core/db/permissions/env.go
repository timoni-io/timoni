package permissions

type EnvPerm uint8

// Enviroment permissions:
//
//go:generate stringer -type=EnvPerm
const (
	Env_View EnvPerm = iota // mozliwosc podglądu danego sr w trybie RO

	Env_Rename         // mozliwosc zmiany nazwy srodowiska
	Env_ManageSchedule // mozliwosc zmiany Schedule w danym srodowisku
	Env_ManageCluster  // mozliwosc zmiany klastra na którym dane środowisko jest uruchomione
	Env_ManageGitOPS   // mozliwosc zmiany sposobu zarządzania środowiskiem webui vs git
	Env_ManageTags     // mozliwosc zmiany Tagów środowiska

	Env_ElementVersionChangeOnly // moge tylko zmieniac wersje elementów, jesli jest w trybie recznej zmiany wersji
	Env_ElementStartStopRestart  // mozliwosc restartowania i zatrzymywania elemtnów/podów
	Env_ElementTerminal          // czy moge otworzyc terminal/cli do srodka PODa
	Env_ElementFullManage        // moge robic wszystkie operacje na elementwch srodowiska

	Env_ViewLogsEvents  // czy moge ogladać logi z  eventów
	Env_ViewLogsBuild   // czy moge ogladać logi z budowania
	Env_ViewLogsRuntime // czy moge ogladać logi z kontenerow

	Env_ViewMembers   // podglad listy uczestnikow paracujacych na danym srodowisku, ale nie wysylany info o konkretnych uprawnieniach tych ludzi
	Env_ManageMembers // pelne zarzadzanie uczestnikami danego srodowiska

	Env_ViewMetrics   // mozliwosc podglądu metryk
	Env_ManageMetrics // mozliwosc zmiany konfiguracji metryk i alertów

	Env_CopyAndViewSecrets // mozliwosc kopiowania secretow

	__Env_Iter // used for iteration, do not remove
)
