package rmq

import (
	"sort"
)

func GetKeys(channels map[string]ChannelSettings) []string {
	var keys = make([]string, 0, len(channels))

	for name := range channels {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}
