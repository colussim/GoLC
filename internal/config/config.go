package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// ----- Structures de configuration -----

// Secret est optionnel : il sera résolu plus tard (env/file/keyring/vault)
type Secret struct {
	Type    string `mapstructure:"type"`   // env|file|keyring|vault (optionnel)
	Name    string `mapstructure:"name"`   // env var (env)
	Path    string `mapstructure:"path"`   // fichier (file)
	Service string `mapstructure:"service"`// keyring
	User    string `mapstructure:"user"`   // keyring
	Addr    string `mapstructure:"addr"`   // vault
	VPath   string `mapstructure:"path"`   // vault secret path
	Field   string `mapstructure:"field"`  // vault field name
}

type PlatformConfig struct {
	Users             string   `mapstructure:"Users"`
	AccessToken       string   `mapstructure:"AccessToken"`
	Organization      string   `mapstructure:"Organization"`
	DevOps            string   `mapstructure:"DevOps"`
	Workspace         string   `mapstructure:"Workspace"` // BitBucket
	Project           string   `mapstructure:"Project"`
	Repos             string   `mapstructure:"Repos"`
	Branch            string   `mapstructure:"Branch"`
	DefaultBranch     bool     `mapstructure:"DefaultBranch"`
	Url               string   `mapstructure:"Url"`
	Apiver            string   `mapstructure:"Apiver"`
	Baseapi           string   `mapstructure:"Baseapi"`
	Protocol          string   `mapstructure:"Protocol"`
	FileExclusion     string   `mapstructure:"FileExclusion"`
	ExtExclusion      []string `mapstructure:"ExtExclusion"`
	ExcludePaths      []string `mapstructure:"ExcludePaths"`
	Period            int      `mapstructure:"Period"`
	Factor            int      `mapstructure:"Factor"`
	Multithreading    bool     `mapstructure:"Multithreading"`
	Stats             bool     `mapstructure:"Stats"`
	Workers           int      `mapstructure:"Workers"`
	NumberWorkerRepos int      `mapstructure:"NumberWorkerRepos"`
	ResultByFile      bool     `mapstructure:"ResultByFile"`
	ResultAll         bool     `mapstructure:"ResultAll"`
	Org               bool     `mapstructure:"Org"`
	Zip               bool     `mapstructure:"Zip"`

	// Spécifique au provider File
	Directory string `mapstructure:"Directory"`
	FileLoad  string `mapstructure:"FileLoad"`

	// Option Secret manager/keystore (facultatif)
	Secret *Secret `mapstructure:"Secret"`
}

type LoggingRotation struct {
	MaxSizeMB   int  `mapstructure:"MaxSizeMB"`
	MaxBackups  int  `mapstructure:"MaxBackups"`
	MaxAgeDays  int  `mapstructure:"MaxAgeDays"`
	Compress    bool `mapstructure:"Compress"`
}

type Logging struct {
	Level    string          `mapstructure:"Level"`   // trace|debug|info|warn|error|fatal
	Format   string          `mapstructure:"Format"`  // json|text (pour la console si --log-json absent; et pour fichier si tu veux)
	ToFile   bool            `mapstructure:"ToFile"`  // active écriture fichier
	Dir      string          `mapstructure:"Dir"`     // dossier des logs (créé si absent)
	File     string          `mapstructure:"File"`    // nom fichier; si vide, généré (golc-YYYYMMDD-HHMM.log)
	Rotation LoggingRotation `mapstructure:"Rotation"`// (facultatif) si tu veux gérer plus tard
}

type Release struct {
	Version string `mapstructure:"Version"`
}

type Config struct {
	Platforms map[string]PlatformConfig `mapstructure:"platforms"`
	Logging   Logging                   `mapstructure:"Logging"`
	Release   Release                   `mapstructure:"Release"`
}

// ----- Chargement Viper -----

// Load lit JSON ou YAML. Si path est vide, tente ./config.yaml|yml|json.
// Expansion ${ENV} dans le fichier, + bind de quelques ENV utiles (tokens).
func Load(path string) (*Config, error) {
	v := viper.New()

	// ENV GOLC_*
	v.SetEnvPrefix("golc")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Fichier
	if path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}
		expanded := []byte(os.ExpandEnv(string(b)))
		switch strings.ToLower(filepath.Ext(path)) {
		case ".json":
			v.SetConfigType("json")
		case ".yml", ".yaml":
			v.SetConfigType("yaml")
		default:
			v.SetConfigType("yaml")
		}
		if err := v.ReadConfig(bytes.NewReader(expanded)); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	} else {
		// Recherche par défaut
		for _, p := range []string{"config.yaml", "config.yml", "config.json"} {
			if b, err := os.ReadFile(p); err == nil {
				expanded := []byte(os.ExpandEnv(string(b)))
				if strings.HasSuffix(p, ".json") {
					v.SetConfigType("json")
				} else {
					v.SetConfigType("yaml")
				}
				if err := v.ReadConfig(bytes.NewReader(expanded)); err != nil {
					return nil, fmt.Errorf("parse %s: %w", p, err)
				}
				break
			}
		}
	}

	// Bind ENV pratiques (si présents, ils écrasent le fichier)
	_ = v.BindEnv("platforms.Github.AccessToken", "GOLC_GITHUB_TOKEN")
	_ = v.BindEnv("platforms.Gitlab.AccessToken", "GOLC_GITLAB_TOKEN")
	_ = v.BindEnv("platforms.BitBucket.AccessToken", "GOLC_BITBUCKET_TOKEN")
	_ = v.BindEnv("platforms.BitBucketSRV.AccessToken", "GOLC_BITBUCKETSRV_TOKEN")
	_ = v.BindEnv("platforms.Azure.AccessToken", "GOLC_AZURE_TOKEN")

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) validate() error {
	if c.Platforms == nil || len(c.Platforms) == 0 {
		return errors.New("no platforms configured (key 'platforms' missing or empty)")
	}
	for name, p := range c.Platforms {
		if p.Factor < 0 {
			return fmt.Errorf("platform %s: Factor must be >= 0", name)
		}
	}
	return nil
}

// ----- Sélection de plateforme -----

// SelectPlatform retourne la plateforme par le NOM DE CLÉ (case-insensitive).
// Ex: "Github", "gitHub", "GITHUB" → clé "Github".
func SelectPlatform(cfg *Config, name string) (*PlatformConfig, error) {
	if cfg == nil {
		return nil, errors.New("nil config")
	}
	want := strings.ToLower(strings.TrimSpace(name))
	for key, p := range cfg.Platforms {
		if strings.ToLower(key) == want {
			pp := p
			return &pp, nil
		}
	}
	return nil, fmt.Errorf("platform %q not found", name)
}

// AutoSelectPlatform renvoie la première plateforme avec DevOps non vide ET (AccessToken non vide OU Secret présent).
func AutoSelectPlatform(cfg *Config) (*PlatformConfig, string, error) {
	if cfg == nil {
		return nil, "", errors.New("nil config")
	}
	for name, p := range cfg.Platforms {
		if strings.TrimSpace(p.DevOps) == "" {
			continue
		}
		if strings.TrimSpace(p.AccessToken) != "" || p.Secret != nil {
			cp := p
			return &cp, name, nil
		}
	}
	return nil, "", errors.New("no suitable platform found (need DevOps + AccessToken or Secret)")
}