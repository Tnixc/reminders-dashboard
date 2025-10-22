package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/pelletier/go-toml/v2"
	"github.com/go-playground/validator/v10"

	"github.com/dlvhdr/reminders-dashboard/v4/internal/utils"
)

const DashDir = "reminders-dashboard"

const ConfigYmlFileName = "config.toml"

const ConfigYamlFileName = "config.toml"

const DEFAULT_XDG_CONFIG_DIRNAME = ".config"

var validate *validator.Validate

type ViewType string

const (
	RemindersView ViewType = "reminders"
)

type SectionConfig struct {
	Title   string
	Filters string
	Limit   *int `toml:"limit,omitempty"`
	Type    *ViewType
}

type RemindersSectionConfig struct {
	Title   string
	Filters string
	Limit   *int                  `toml:"limit,omitempty"`
	Layout  RemindersLayoutConfig `toml:"layout,omitempty"`
	Type    *ViewType
}

type PreviewConfig struct {
	Open  bool
	Width int
}

type ColumnConfig struct {
	Width  *int  `toml:"width,omitempty"  validate:"omitempty,gt=0"`
	Hidden *bool `toml:"hidden,omitempty"`
}

type RemindersLayoutConfig struct {
	Title       ColumnConfig `toml:"title,omitempty"`
	List        ColumnConfig `toml:"list,omitempty"`
	DueIn       ColumnConfig `toml:"dueIn,omitempty"`
	Date        ColumnConfig `toml:"date,omitempty"`
	Priority    ColumnConfig `toml:"priority,omitempty"`
	Completed   ColumnConfig `toml:"completed,omitempty"`
}

type LayoutConfig struct {
	Reminders RemindersLayoutConfig `toml:"reminders,omitempty"`
}

type Defaults struct {
	Preview                PreviewConfig `toml:"preview"`
	RemindersLimit         int           `toml:"remindersLimit"`
	View                   ViewType      `toml:"view"`
	Layout                 LayoutConfig  `toml:"layout,omitempty"`
	RefetchIntervalMinutes int           `toml:"refetchIntervalMinutes,omitempty"`
	DateFormat             string        `toml:"dateFormat,omitempty"`
}

type RepoConfig struct {
	BranchesRefetchIntervalSeconds int `toml:"branchesRefetchIntervalSeconds,omitempty"`
	PrsRefetchIntervalSeconds      int `toml:"prsRefetchIntervalSeconds,omitempty"`
}

type Keybinding struct {
	Key     string `toml:"key"`
	Command string `toml:"command"`
	Builtin string `toml:"builtin"`
	Name    string `toml:"name,omitempty"`
}

func (kb Keybinding) NewBinding(previous *key.Binding) key.Binding {
	helpDesc := ""
	if previous != nil {
		helpDesc = previous.Help().Desc
	}

	if kb.Name != "" {
		helpDesc = kb.Name
	}

	return key.NewBinding(
		key.WithKeys(kb.Key),
		key.WithHelp(kb.Key, helpDesc),
	)
}

type Keybindings struct {
	Universal  []Keybinding `toml:"universal"`
	Reminders  []Keybinding `toml:"reminders"`
}

type Pager struct {
	Diff string `toml:"diff"`
}

type HexColor string



type ColorThemeText struct {
	Primary   HexColor `toml:"primary"   validate:"omitempty,hexcolor"`
	Secondary HexColor `toml:"secondary" validate:"omitempty,hexcolor"`
	Inverted  HexColor `toml:"inverted"  validate:"omitempty,hexcolor"`
	Faint     HexColor `toml:"faint"     validate:"omitempty,hexcolor"`
	Warning   HexColor `toml:"warning"   validate:"omitempty,hexcolor"`
	Success   HexColor `toml:"success"   validate:"omitempty,hexcolor"`
	Error     HexColor `toml:"error"     validate:"omitempty,hexcolor"`
}

type ColorThemeBorder struct {
	Primary   HexColor `toml:"primary"   validate:"omitempty,hexcolor"`
	Secondary HexColor `toml:"secondary" validate:"omitempty,hexcolor"`
	Faint     HexColor `toml:"faint"     validate:"omitempty,hexcolor"`
}

type ColorThemeBackground struct {
	Selected HexColor `toml:"selected" validate:"omitempty,hexcolor"`
}

type ColorTheme struct {
	Text       ColorThemeText       `toml:"text"       validate:"required"`
	Background ColorThemeBackground `toml:"background" validate:"required"`
	Border     ColorThemeBorder     `toml:"border"     validate:"required"`
}

type ColorThemeConfig struct {
	Inline ColorTheme `toml:",inline"`
}



type TableUIThemeConfig struct {
	ShowSeparator bool `toml:"showSeparator" default:"true"`
	Compact       bool `toml:"compact" default:"false"`
}

type UIThemeConfig struct {
	SectionsShowCount bool               `toml:"sectionsShowCount" default:"true"`
	Table             TableUIThemeConfig `toml:"table"`
}

type ThemeConfig struct {
	Ui     UIThemeConfig     `toml:"ui,omitempty"     validate:"omitempty"`
	Colors *ColorThemeConfig `toml:"colors,omitempty" validate:"omitempty"`
}

type Config struct {
	RemindersSections      []RemindersSectionConfig `toml:"remindersSections"`
	Defaults               Defaults                 `toml:"defaults"`
	Keybindings            Keybindings              `toml:"keybindings"`
	Theme                  *ThemeConfig             `toml:"theme,omitempty" validate:"omitempty"`
	ListColors             map[string]string        `toml:"listColors"`
	ConfirmQuit            bool                     `toml:"confirmQuit"`
	SmartFilteringAtLaunch bool                     `toml:"smartFilteringAtLaunch" default:"true"`
}

type configError struct {
	configDir string
	parser    ConfigParser
	err       error
}

type ConfigParser struct{}

func (parser ConfigParser) getDefaultConfig() Config {
	return Config{
		Defaults: Defaults{
			Preview: PreviewConfig{
				Open:  true,
				Width: 50,
			},
			RemindersLimit:         20,
			View:                   RemindersView,
			RefetchIntervalMinutes: 0,
			Layout: LayoutConfig{
				Reminders: RemindersLayoutConfig{
					Title: ColumnConfig{
						Width: utils.IntPtr(30),
					},
					List: ColumnConfig{
						Width: utils.IntPtr(20),
					},
					DueIn: ColumnConfig{
						Width: utils.IntPtr(15),
					},
					Date: ColumnConfig{
						Width: utils.IntPtr(lipgloss.Width("2023-12-31 15:04")),
					},
					Priority: ColumnConfig{
						Width: utils.IntPtr(10),
					},
					Completed: ColumnConfig{
						Width: utils.IntPtr(10),
					},
				},
			},
		},
		RemindersSections: []RemindersSectionConfig{
			{
				Title:   "All Reminders",
				Filters: "",
			},
		},
		Keybindings: Keybindings{
			Universal: []Keybinding{},
			Reminders: []Keybinding{},
		},
		Theme: &ThemeConfig{
			Ui: UIThemeConfig{
				SectionsShowCount: true,
				Table: TableUIThemeConfig{
					ShowSeparator: true,
					Compact:       false,
				},
			},
		},
		ListColors:             map[string]string{},
		ConfirmQuit:            false,
		SmartFilteringAtLaunch: true,
	}
}

func (parser ConfigParser) getDefaultConfigContents() string {
	defaultConfig := parser.getDefaultConfig()
	tomlBytes, _ := toml.Marshal(defaultConfig)

	return string(tomlBytes)
}

func (e configError) Error() string {
	return fmt.Sprintf(
		`Couldn't find a config.toml configuration file.
Create one under: %s

Example of a config.toml file:
%s

For more info, go to https://github.com/dlvhdr/reminders-dashboard
press q to exit.

Original error: %v`,
		path.Join(e.configDir, DashDir, ConfigYmlFileName),
		string(e.parser.getDefaultConfigContents()),
		e.err,
	)
}

func (parser ConfigParser) writeDefaultConfigContents(
	newConfigFile *os.File,
) error {
	_, err := newConfigFile.WriteString(parser.getDefaultConfigContents())
	if err != nil {
		return err
	}

	return nil
}

func (parser ConfigParser) createConfigFileIfMissing(
	configFilePath string,
) error {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		newConfigFile, err := os.OpenFile(
			configFilePath,
			os.O_RDWR|os.O_CREATE|os.O_EXCL,
			0o666,
		)
		if err != nil {
			return err
		}

		defer newConfigFile.Close()
		return parser.writeDefaultConfigContents(newConfigFile)
	}

	return nil
}

func (parser ConfigParser) getDefaultConfigFileOrCreateIfMissing(repoPath string) (string, error) {
	var configFilePath string
	ghDashConfig := os.Getenv("GH_DASH_CONFIG")

	// First try GH_DASH_CONFIG
	if ghDashConfig != "" {
		configFilePath = ghDashConfig
	}

	// Then fallback to global config
	if configFilePath == "" {
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			configDir = filepath.Join(homeDir, DEFAULT_XDG_CONFIG_DIRNAME)
		}

		dashConfigDir := filepath.Join(configDir, DashDir)
		configFilePath = filepath.Join(dashConfigDir, ConfigYmlFileName)
	}

	// Ensure directory exists before attempting to create file
	configDir := filepath.Dir(configFilePath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err = os.MkdirAll(configDir, os.ModePerm); err != nil {
			return "", configError{
				parser:    parser,
				configDir: configDir,
				err:       err,
			}
		}
	}

	if err := parser.createConfigFileIfMissing(configFilePath); err != nil {
		return "", configError{parser: parser, configDir: configDir, err: err}
	}

	return configFilePath, nil
}

type parsingError struct {
	err error
}

func (e parsingError) Error() string {
	return fmt.Sprintf("failed parsing config.toml: %v", e.err)
}

func (parser ConfigParser) readConfigFile(path string) (Config, error) {
	config := parser.getDefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return config, configError{parser: parser, configDir: path, err: err}
	}

	err = toml.Unmarshal([]byte(data), &config)
	if err != nil {
		return config, err
	}

	err = validate.Struct(config)
	return config, err
}

func initParser() ConfigParser {
	validate = validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.Split(fld.Tag.Get("toml"), ",")[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return ConfigParser{}
}

type Location struct {
	RepoPath string
	// Config passed with explicit --config flag
	ConfigFlag string
}

func ParseConfig(location Location) (Config, error) {
	parser := initParser()

	var config Config
	var err error
	var configFilePath string

	// give priority to `--config` flag
	if location.ConfigFlag == "" {
		configFilePath, err = parser.getDefaultConfigFileOrCreateIfMissing(location.RepoPath)
		if err != nil {
			return config, parsingError{err: err}
		}
	} else {
		configFilePath = location.ConfigFlag
	}

	config, err = parser.readConfigFile(configFilePath)
	if err != nil {
		return config, parsingError{err: err}
	}

	return config, nil
}
