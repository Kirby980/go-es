package builder

import "sync"

var mapPool = sync.Pool{
	New: func() any {
		return make(map[string]any)
	},
}

func getMap() map[string]any {
	return mapPool.Get().(map[string]any)
}

func putMap(m map[string]any) {
	for k := range m {
		delete(m, k)
	}
	mapPool.Put(m)
}
