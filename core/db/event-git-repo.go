package db

import (
	"core/db2"
	"core/db2/fp"
	"fmt"
	"runtime/debug"
	"strings"

	log "lib/tlog"
)

const PANEL_URL = "https://%s/env/%s/%s/overview" // 1: Domain, 2: Project.Name, 3: App.ID

// 1: Domain, 2: Project.Name
const (
	CONTROL_PANEL_URL = "https://%s/project/%s/control"
	LIMITS_URL        = "https://%s/project/%s/limits"
	COMMITS_URL       = "https://%s/code/%s/commits"
)

func EventMemberAdded(
	project *GitRepoS,
	member UserS, 
	user UserS,
) {
	log.Info("Member added", log.Vars{
		"event":      "true",
		"project":    project.Name,
		"user":       user.Email,
		"member":     member.Email,
		"memberRole": project.Operators[member.Email],
	})

	if project.Notifications.Member == nil {
		project.Notifications.Member = map[string]bool{}
	}

}

func EventMemberRemoved(
	project *GitRepoS, 
	memberEmail string,
	memberRole string,
	user UserS, 
) {
	log.Info("Member removed", log.Vars{
		"event":      "true",
		"project":    project.Name,
		"user":       user.Email,
		"member":     memberEmail,
		"memberRole": memberRole,
	})

	if project.Notifications.Member == nil {
		project.Notifications.Member = map[string]bool{}
	}

}

func EventMemberRoleChanged(
	project *GitRepoS, 
	memberEmail string,
	oldRole string,
	newRole string, 
	user UserS, 
) {
	log.Info("Member role changed", log.Vars{
		"event":         "true",
		"project":       project.Name,
		"user":          user.Email,
		"member":        memberEmail,
		"memberRoleOld": oldRole,
		"memberRoleNew": newRole,
	})

	if project.Notifications.Member == nil {
		project.Notifications.Member = map[string]bool{}
	}
}


func EventLimitChanged(
	project *GitRepoS, 
	limitName string,
	limitValueOld int,
	limitValueNew int,
	user UserS, 
) {
	log.Info("Limit changed", log.Vars{
		"event":         "true",
		"project":       project.Name,
		"user":          user.Email,
		"limitName":     limitName,
		"limitValueOld": limitValueOld,
		"limitValueNew": limitValueNew,
	})

	if project.Notifications.Limit == nil {
		project.Notifications.Limit = map[string]bool{}
	}
}

func EventProjectSetingsChanged(
	project *GitRepoS, 
	user UserS, 
) {
	log.Info("Setings changed", log.Vars{
		"event":   "true",
		"user":    user.Email,
		"project": project.Name,
	})

	if project.Notifications.Settings == nil {
		project.Notifications.Settings = map[string]bool{}
	}

}

func EventCodeBranchCreated(
	project *GitRepoS, 
	gitBranch string, // jakiego branch dotyczy powiadomienie/zdarzenie
	user UserS, 
) {
	log.Info("Branch created", log.Vars{
		"event":   "true",
		"project": project.Name,
		"user":    user.Email,
		"branch":  gitBranch,
	})

	if project.Notifications.CodeBranche == nil {
		project.Notifications.CodeBranche = map[string]bool{}
	}
}

func EventCodeCommit(
	project *GitRepoS, 
	gitBranch string, // jakiego branch dotyczy powiadomienie/zdarzenie
	user UserS, 
) {
	log.Info("New commit", log.Vars{
		"event":     "true",
		"project":   project.Name,
		"user":      user.Email,
		"gitBranch": gitBranch,
	})

	if project.Notifications.CodeCommit == nil {
		project.Notifications.CodeCommit = map[string]map[string]bool{}
	}
	if project.Notifications.CodeCommit[gitBranch] == nil {
		project.Notifications.CodeCommit[gitBranch] = map[string]bool{}
	}
}

func EventRemoteGitAccessDenied(project *GitRepoS) {
	log.Error("Remote git repo access denied", log.Vars{
		"event":     "true",
		"project":   project.Name,
		"remoteURL": project.RemoteURL,
	})
}

func PanicHandler() {
	if err := recover(); err != nil {
		fmt.Println("Panic:", err)
		_, stack, _ := strings.Cut(string(debug.Stack()), "panic")
		fmt.Println("Stack:", stack)

		if db2.TheDomain.Name() != "local" {
			fp.SendEmail("lw@nri.pl", "PANIC - Core - "+db2.TheDomain.Name(),
				fmt.Sprintf("Panic: %s\nStack: %s\n", err, stack))
		}
	}
}

func PanicHandlerCB(fn func()) bool {
	if err := recover(); err != nil {
		fmt.Println("Panic:", err)
		_, stack, _ := strings.Cut(string(debug.Stack()), "panic")
		fmt.Println("Stack:", stack)

		if db2.TheDomain.Name() != "local" {
			fp.SendEmail("lw@nri.pl", "PANIC - Core - "+db2.TheDomain.Name(),
				fmt.Sprintf("Panic: %s\nStack: %s\n", err, stack))
		}

		fn()
		return true
	}

	return false
}
