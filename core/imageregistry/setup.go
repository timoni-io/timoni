package imageregistry

import "core/modulestate"

func Setup() {
	modulestate.StatusByModulesAdd("image-registry", Check)
}
