package rmq

import (
	core "bb_core"
	"fmt"
	"sort"
)

// func sortByOrder(channels map[string]core.ChannelSettings) {

// }

func getKeys(channels map[string]core.ChannelSettings) []string {
	var keys = make([]string, 0, len(channels))

	for name := range channels {
		keys = append(keys, name)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		order1 := channels[keys[i]].Order
		order2 := channels[keys[j]].Order
		return order1 < order2
	})
	return keys
}

func GetRabbitUrl(settings core.RabbitMQ) string {
	template := "%s://%s:%s@%s:%d"
	conn := settings.Connection
	protocol, hostname, username, password, port :=
		conn.Protocol,
		conn.Hostname,
		conn.Username,
		conn.Password,
		conn.Port

	// reflectConnection := reflect.TypeOf(settings)
	// getConfigValue(reflectConnection, &protocol, "Protocol")
	// getConfigValue(reflectConnection, &hostname, "Hostname")
	// getConfigValue(reflectConnection, &username, "Username")
	// getConfigValue(reflectConnection, &password, "Password")
	// getConfigIntValue(reflectConnection, &port, "Port")

	return fmt.Sprintf(template, protocol, username, password, hostname, port)
}
