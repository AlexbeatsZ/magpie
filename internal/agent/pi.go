package agent

// Pi's enabledModels (settings.json) is its Ctrl+P cycle and /scoped-models
// list. While it names anything, a new session starts on defaultModel only
// when the list takes that model in, and on the list's first model when it
// doesn't — so a model magpie picks outside it would never be used. The list
// is the user's: magpie adds the model it picks to it, as Pi's own model
// picker does, and never makes one where there is none.

// piLevels are the thinking levels a pattern may end in (provider/model:high).
var piLevels = []string{"off", "minimal", "low", "medium", "high", "xhigh", "max"}
