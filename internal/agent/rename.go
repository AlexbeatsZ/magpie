package agent

import (
	"github.com/yetone/magpie/internal/provider"
)

// RenameProvider gives a provider another id (provider.Rename) and moves
// the agents on one of its models to the same model by the new id, each
// spelled the way the agent spells it, and what else in their configs names
// it (RenameRefs). It answers the agents it moved.
func RenameProvider(from, to string) ([]string, error) {
	return nil, provider.Rename(from, to)
}

// RenameGroup gives a routing group another id (provider.RenameGroup) and
// moves the agents on it to group/<to>, as RenameProvider does a
// provider's: left on group/<from>, each agent named a group magpie no
// longer has, and the user had to pick the group again in every agent
// (lucis0116 on X). It answers the agents it moved.
func RenameGroup(from, to string) ([]string, error) {
	return nil, provider.RenameGroup(from, to)
}
