package permissions

import (
	"lib/utils/bitmap"
	"strings"
)

var (
	DefaultGroup PermGroup
)

type PermGroup struct {
	Global Mask

	// Env Tag -> Mask

	Envs     map[string]Mask
	GitRepos map[string]Mask
}

func (pg PermGroup) ToFrontPerm() FrontPerm {
	front := FrontPerm{
		Global:   map[string]PermExplained{},
		Envs:     map[string]map[string]PermExplained{},
		GitRepos: map[string]map[string]PermExplained{},
	}

	// Global
	for i := GlobPerm(0); i < __Glob_Iter; i++ {
		front.Global[i.String()] = PermExplained{
			Index: uint8(i),
			IsSet: bitmap.GetBit(pg.Global, i),
		}
	}

	// Envs
	for tag, mask := range pg.Envs {
		m := map[string]PermExplained{}
		for i := EnvPerm(0); i < __Env_Iter; i++ {
			m[i.String()] = PermExplained{
				Index: uint8(i),
				IsSet: bitmap.GetBit(mask, i),
			}
		}
		front.Envs[tag] = m
	}

	// GitRepos
	for tag, mask := range pg.GitRepos {
		m := map[string]PermExplained{}
		for i := RepoPerm(0); i < __Repo_Iter; i++ {
			m[i.String()] = PermExplained{
				Index: uint8(i),
				IsSet: bitmap.GetBit(mask, i),
			}
		}
		front.GitRepos[tag] = m
	}
	return front
}

type Mask uint32

func (pm *Mask) Join(m Mask) {
	bitmap.Join(pm, m)
}

func FromMap(m map[uint8]bool) (pm Mask) {
	for idx, val := range m {
		if val {
			bitmap.SetBit(&pm, idx)
		}
	}
	return
}

type FrontPerm struct {
	Global map[string]PermExplained

	// Tag -> Permission Name -> PermExplained

	Envs     map[string]map[string]PermExplained
	GitRepos map[string]map[string]PermExplained
}

type PermExplained struct {
	Index uint8
	IsSet bool
}

func (fp FrontPerm) Env(id string, tags []string) map[string]PermExplained {

	for selectorList, prems := range fp.Envs {
		for _, selector := range strings.Split(selectorList, ";") {
			if selector == "id:"+id {
				return prems
			}
		}
	}

	return nil
}
