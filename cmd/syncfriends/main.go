package main

import (
	"github.com/loks666/FlyAI/initialization"
	"github.com/loks666/FlyAI/tasks/friends"
)

// syncfriends performs a one-shot contact and group-member synchronization
// against the configured wxhelper-compatible bridge.
func main() {
	initialization.InitConfig()
	friends.Sync()
}
