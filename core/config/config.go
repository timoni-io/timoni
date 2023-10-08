package config

import (
	"path/filepath"

	"github.com/lukx33/lwhelper"
)

var (
	GitTag    = "???"
	CommitSHA = "???"

	InstallationID = lwhelper.GetEnv("LicenseKey", "")
	FocalPointAddr = lwhelper.GetEnv("FocalPointAddr", "fp.timoni.io:53")
	DataPath       = lwhelper.GetEnv("DataPath", "/data")
	ModulesPath    = lwhelper.GetEnv("ModulesPath", "/modules")
	WebPublicPath  = lwhelper.GetEnv("WebPublicPath", "/public")

	KubeConfigFilePath = filepath.Join(DataPath(), "kubeconfig.yaml")
	GitStatsPath       = filepath.Join(DataPath(), "git-stats")
	GitRemotePath      = filepath.Join(DataPath(), "git-remote")
	GitLocalPath       = filepath.Join(DataPath(), "git-local")
)
