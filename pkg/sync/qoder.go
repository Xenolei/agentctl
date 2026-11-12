package sync

import (
	"os"
	"path/filepath"

	"github.com/iheanyi/agentctl/pkg/mcp"
	"github.com/iheanyi/agentctl/pkg/skill"
)

// QoderAdapter syncs configuration to Qoder CN (Alibaba's AI IDE, China version).
// Skills live in ~/.qoder-cn/skills/. MCP servers use ~/.qoder-cn/mcp.json
// (standard "mcpServers" JSON). Verified against a Qoder CN install: the UI only
// lists ~/.qoder-cn/skills, and the MCP config file is mcp.json.
type QoderAdapter struct{}

func init() {
	Register(&QoderAdapter{})
}

func (a *QoderAdapter) Name() string {
	return "qoder"
}

func (a *QoderAdapter) baseDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".qoder-cn")
}

func (a *QoderAdapter) Detect() (bool, error) {
	dir := a.baseDir()
	if dir == "" {
		return false, nil
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func (a *QoderAdapter) ConfigPath() string {
	return filepath.Join(a.baseDir(), "mcp.json")
}

func (a *QoderAdapter) SupportedResources() []ResourceType {
	return []ResourceType{ResourceMCP, ResourceSkills}
}

// ReadServers reads MCP server configurations from Qoder.
func (a *QoderAdapter) ReadServers() ([]*mcp.Server, error) {
	helper := NewJSONConfigHelper(a.ConfigPath())
	raw, err := helper.LoadRaw()
	if err != nil {
		return nil, err
	}

	mcpServers, _ := GetMCPServersSection(raw, "mcpServers")
	return ServersFromMCPSection(mcpServers), nil
}

// WriteServers writes MCP server configurations to Qoder.
func (a *QoderAdapter) WriteServers(servers []*mcp.Server) error {
	helper := NewJSONConfigHelper(a.ConfigPath())
	raw, err := helper.LoadRaw()
	if err != nil {
		return err
	}

	mcpServers, _ := GetMCPServersSection(raw, "mcpServers")
	RemoveManagedServers(mcpServers)

	for _, server := range servers {
		name := GetServerName(server)
		if name == "" {
			continue
		}
		mcpServers[name] = ServerToRawMap(server)
	}

	raw["mcpServers"] = mcpServers
	return helper.SaveRaw(raw)
}

// skillsDir returns Qoder's skills directory.
func (a *QoderAdapter) skillsDir() string {
	return filepath.Join(a.baseDir(), "skills")
}

// ReadSkills reads installed skills from Qoder's skills directory.
func (a *QoderAdapter) ReadSkills() ([]*skill.Skill, error) {
	return ReadSkillsFromDir(a.skillsDir())
}

// WriteSkills writes skills to Qoder's skills directory.
func (a *QoderAdapter) WriteSkills(skills []*skill.Skill) error {
	return WriteSkillsToDir(a.skillsDir(), skills)
}
