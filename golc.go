package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/colussim/GoLC/assets"
	appcfg "github.com/colussim/GoLC/internal/config"
)

var (
	flagConfig    string
	flagDevOps    string
	flagDocker    bool
	flagFast      bool
	flagLanguages bool
	flagVersion   bool
	flagLogFile   string
	flagLogJSON   bool

	log = logrus.New()
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "golc --devops <BitBucketSRV|BitBucket|Github|Gitlab|Azure|File> [OPTIONS]",
		Short: "GoLC – count physical LoC across many repos/providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			bootstrapLogger()

			// Démarrage
			log.Info("🚀 Starting GoLC")

			// --version : affiche et sort
			if flagVersion {
				cfg, _ := appcfg.Load(flagConfig)
				ver := "dev"
				if cfg != nil && cfg.Release.Version != "" {
					ver = cfg.Release.Version
				}
				log.Infof("ℹ️  Version: %s", ver)
				fmt.Println(ver)
				return nil
			}

			// --languages : liste les langages supportés
			if flagLanguages {
				log.Info("📚 Displaying supported languages")
				displayLanguages()
				log.Info("✅ Languages displayed")
				return nil
			}

			if strings.TrimSpace(flagDevOps) == "" {
				logrus.Fatalf("❌ Missing required flag: --devops <BitBucketSRV|BitBucket|Github|Gitlab|Azure|File>")
			}

			// Chargement config
			cfg, err := appcfg.Load(flagConfig)
			if err != nil {
				logrus.Fatalf("❌ Failed to load config: %s", err)
			}
			log.Info("✅ Configuration loaded successfully")

			// Configurer logger (fichier + niveau)
			if err := configureLoggerFromConfig(cfg); err != nil {
				log.Warnf("⚠️  File logging disabled (init failed): %v", err)
			} else if flagLogFile != "" || cfg.Logging.ToFile {
				log.Info("📝 File logging enabled")
			}

			// Sélection plateforme
			plat, err := appcfg.SelectPlatform(cfg, flagDevOps)
			if err != nil {
				var keys []string
				for k := range cfg.Platforms {
					keys = append(keys, k)
				}
				logrus.Fatalf("❌ Platform %q not found. Configured platforms: %s", flagDevOps, strings.Join(keys, ", "))
			}
			log.WithFields(logrus.Fields{
				"devops": plat.DevOps,
				"url":    plat.Url,
				"org":    plat.Organization,
			}).Info("✅ Platform selected")

			// Vérifs d’options
			if flagFast && !equalsFold(plat.DevOps, "github") {
				log.Warn("⚠️  --fast is only meaningful for Github; ignored")
			}
			if flagDocker {
				log.Info("🐳 Docker mode requested (handled later in pipeline)")
			}

			// Affiche la plateforme retenue (JSON pretty pour debug)
			out, _ := json.MarshalIndent(plat, "", "  ")
			fmt.Println(string(out))
			log.Info("✅ Platform details printed")

			// TODO: brancher le pipeline de scan ici
			log.Info("🏁 GoLC initialized successfully (ready to scan)")
			return nil
		},
	}

	// Flags
	rootCmd.Flags().StringVar(&flagConfig, "config", "", "path to config file (yaml|yml|json). Defaults to ./config.{yaml|yml|json}")
	rootCmd.Flags().StringVar(&flagDevOps, "devops", "", "Specify the DevOps platform: BitBucketSRV|BitBucket|Github|Gitlab|Azure|File")
	rootCmd.Flags().BoolVar(&flagDocker, "docker", false, "Run in Docker mode")
	rootCmd.Flags().BoolVar(&flagFast, "fast", false, "Enable fast mode (only for Github)")
	rootCmd.Flags().BoolVar(&flagLanguages, "languages", false, "Show all supported languages")
	rootCmd.Flags().BoolVar(&flagVersion, "version", false, "Show version")
	rootCmd.Flags().StringVar(&flagLogFile, "log-file", "", "also write logs to this file (in addition to stdout)")
	rootCmd.Flags().BoolVar(&flagLogJSON, "log-json", false, "use JSON formatter for console logs")

	rootCmd.SilenceUsage = true

	if err := rootCmd.Execute(); err != nil {
		logrus.Fatalf("❌ %v", err)
	}
}

// -------- Logging ----------

func bootstrapLogger() {
	log.SetOutput(os.Stdout)
	if flagLogJSON {
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "15:04:05",
		})
	}
	log.SetLevel(logrus.InfoLevel)
}

func configureLoggerFromConfig(cfg *appcfg.Config) error {
	// Niveau
	switch strings.ToLower(strings.TrimSpace(cfg.Logging.Level)) {
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "warn", "warning":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}

	// Console JSON si demandé par flag
	if flagLogJSON {
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	}

	// Fichier : priorité au flag --log-file
	if flagLogFile != "" {
		if err := ensureDir(filepath.Dir(flagLogFile)); err != nil {
			return err
		}
		f, err := os.OpenFile(flagLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		log.SetOutput(io.MultiWriter(os.Stdout, f))
		log.Infof("🗂️  Logging to file: %s", flagLogFile)
		return nil
	}

	// Sinon, configuration
	if !cfg.Logging.ToFile {
		log.SetOutput(os.Stdout)
		return nil
	}
	dir := cfg.Logging.Dir
	if strings.TrimSpace(dir) == "" {
		dir = "Logs"
	}
	if err := ensureDir(dir); err != nil {
		return err
	}
	filename := cfg.Logging.File
	if strings.TrimSpace(filename) == "" {
		filename = fmt.Sprintf("golc-%s.log", time.Now().Format("20060102-150405"))
	}
	path := filepath.Join(dir, filename)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	log.SetOutput(io.MultiWriter(os.Stdout, f))
	log.Infof("🗂️  Logging to file: %s", path)
	return nil
}

func ensureDir(d string) error {
	if d == "" || d == "." {
		return nil
	}
	return os.MkdirAll(d, 0o755)
}

func equalsFold(a, b string) bool { return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b)) }

// -------- Languages ----------

func displayLanguages() {
	fmt.Printf("%-18s | %-78s | %-15s | %s\n", "Language", "Extensions", "Single Comments", "Multi Line Comments")
	fmt.Println("-------------------+--------------------------------------------------------------------------------+-----------------+--------------------")

	for lang, cfg := range assets.Languages {
		extensions := strings.Join(cfg.Extensions, ", ")
		singleComments := strings.Join(cfg.LineComments, ", ")

		multiLineComments := ""
		for _, pair := range cfg.MultiLineComments {
			for _, tok := range pair {
				if multiLineComments != "" {
					multiLineComments += " "
				}
				multiLineComments += tok
			}
		}
		fmt.Printf("%-18s | %-78s | %-15s | %s\n", lang, extensions, singleComments, multiLineComments)
	}
}