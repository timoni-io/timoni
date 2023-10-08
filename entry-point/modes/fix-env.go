package modes

import (
	"entry-point/global"
	"lib/tlog"
	"os"
	"strings"
)

// ApplyVariablesOnFiles replaces templates {{variable}} in specified files with actual environment variables.
func ApplyVariablesOnFiles() {
	for _, filePath := range global.ApplyVariablesOnFiles {
		if filePathFixed := strings.TrimSpace(filePath); filePathFixed != "" {
			applyEnvsOnFile(filePathFixed, global.InitialEnvs)
		}
	}
}

func applyEnvsOnFile(filePath string, envMap map[string]string) {
	tlog.Info("Applying env variables on file: " + filePath)

	buf, err := os.ReadFile(filePath)
	if err != nil {
		tlog.Error(err.Error())
		return
	}

	s := string(buf)
	for k, v := range envMap {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}

	if err = os.WriteFile(filePath, []byte(s), 0644); err != nil {
		tlog.Error(err.Error())
	}
}
