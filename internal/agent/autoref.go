package agent

import ()

// A group magpie found is listed as auto/<model> (provider.Group.Ref); it
// was group/auto-<slug>, and agents magpie set to it then still say so.
// The gateway serves that id as it did, but the lists magpie writes into
// an agent's files name the group by its new id only, so an agent that
// checks its model against its list (OpenCode) would find it gone. On its
// next write of the agent's lists (SyncCatalog), magpie moves each field
// it set to the old id onto the new one; a field the user set themselves,
// or changed since magpie set it, stays as they left it. So does a field
// magpie set to an auto/<model> spelt as the group is listed no longer
// (auto/glm-5-3-flash, now auto/glm-5.3-flash: the spelling follows
// models.dev's, Group.Spelled), which the gateway serves as ever.
