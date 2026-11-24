//go:build !darwin

package sysint

// FinderKind identifies which Finder Quick Action workflow is managed.
type FinderKind string

const (
	FinderKindFolder FinderKind = "folder"
	FinderKindFile   FinderKind = "file"
)

// FinderMenuOptions configures Finder context menu integration.
type FinderMenuOptions struct {
	Folder                 bool
	File                   bool
	OpenCurrentFolderTitle string
	OpenFolderTitle        string
	OpenFileTitle          string
}

// Manager is a no-op stub on non-macOS platforms.
type Manager struct {
	AppBundle            string
	CLIDirs              []string
	ServicesDir          string
	FinderSyncConfigPath string
	SkipPBS              bool
	Handlers             HandlerBackend
}

// DefaultManager returns a stub Manager on non-macOS platforms.
func DefaultManager() *Manager {
	return &Manager{}
}

// ResolveAppBundle returns an empty string on non-macOS platforms.
func ResolveAppBundle() string { return "" }

func (m *Manager) Supported() bool { return false }

func (m *Manager) Status() Status { return Status{} }

func (m *Manager) CLIStatus() (string, bool) { return "", false }

func (m *Manager) InstallCLI() (string, error) { return "", ErrUnsupported }

func (m *Manager) UninstallCLI() error { return nil }

func (m *Manager) HasFinderWorkflow(FinderKind) bool { return false }

func (m *Manager) SyncFinderMenu(FinderMenuOptions) error { return ErrUnsupported }

func (m *Manager) UninstallAllFinderWorkflows() error { return nil }

func (m *Manager) AssociateExt(string, map[string]string) (string, error) { return "", ErrUnsupported }

func (m *Manager) RestoreExt(string, map[string]string) error { return nil }

func (m *Manager) GetDefaultHandler(string) (string, error) { return "", ErrUnsupported }
