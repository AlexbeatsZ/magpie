package agent

import (
	"path/filepath"
	"strings"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/proc"
)

func findNodeModules(bin string, pkgs []string) []string {
	path, err := proc.LookPath(bin)
	if err != nil {
		return nil
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		real = path
	}
	slash := filepath.ToSlash(real)
	for _, pkg := range pkgs {
		modules := ""
		if i := strings.Index(slash, "/node_modules/"+pkg+"/"); i >= 0 {
			modules = filepath.FromSlash(slash[:i] + "/node_modules")
		} else if d := filepath.Join(filepath.Dir(path), "node_modules"); isDir(filepath.Join(d, filepath.FromSlash(pkg))) {
			modules = d // npm's shim on Windows sits beside its node_modules
		}
		if modules != "" {
			return []string{filepath.Join(modules, filepath.FromSlash(pkg), "node_modules"), modules}
		}
	}
	return nil
}

// registryModel is a model as Pi's and omp's registries list it.
type registryModel struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Type    string   `json:"type"` // Pi: chat, image, classifier
	Kind    string   `json:"kind"` // omp: "" for chat, image, embedding…
	Input   []string `json:"input"`
	Context int      `json:"contextWindow"`
	Output  int      `json:"maxTokens"`
}

func (m registryModel) model(provider string) catalog.Model {
	name := m.Name
	if name == "" {
		name = m.ID
	}
	images := false
	for _, in := range m.Input {
		images = images || in == "image"
	}
	return catalog.Model{ID: m.ID, Name: name, Provider: provider, Images: images, Context: m.Context, Output: m.Output}
}
