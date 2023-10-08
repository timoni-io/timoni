package modulestate

import (
	"core/db2"
	"fmt"
	"lib/utils/maps"
	"time"
)

var (
	checks = maps.SafeMap[string, func() (db2.StateT, string)]{} // key= module-name/check-name
)

func Loop() {

	for {
		for _, mod := range checks.Values() {
			go func(mod func() (db2.StateT, string)) {
				state, message := mod()
				if state != db2.State_ready {
					fmt.Println("modulestate:", state, message)
				}
			}(mod)
		}
		time.Sleep(34 * time.Second)
	}
}

func StatusByModulesAdd(name string, fn func() (db2.StateT, string)) {
	checks.Set(name, fn)
}
