package rmq

import (
	core "bb_core"
	"fmt"
	"sort"
)

func getKeys(channels map[string]core.ChannelSettings) []string {
	var keys = make([]string, 0, len(channels))

	for name := range channels {
		keys = append(keys, name)
	}
	sort.Strings(keys)
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
