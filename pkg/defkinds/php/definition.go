package php

import (
	"fmt"
	"path"
	"strings"

	"github.com/NiR-/zbuild/pkg/builddef"
	"github.com/NiR-/zbuild/pkg/defkinds/webserver"
	"github.com/NiR-/zbuild/pkg/llbutils"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
)

// defaultDefinition returns a Definition with all its fields initialized with
// default values.
func defaultDefinition() Definition {
	fpm := true
	healthcheck := false
	infer := true
	isDev := true
	isNotDev := false

	return Definition{
		BaseStage: Stage{
			ExternalFiles:  []llbutils.ExternalFile{},
			SystemPackages: map[string]string{},
			FPM:            &fpm,
			Extensions:     map[string]string{},
			GlobalDeps:     map[string]string{},
			ComposerDumpFlags: &ComposerDumpFlags{
				ClassmapAuthoritative: true,
			},
			Sources:      []string{},
			Integrations: []string{},
			StatefulDirs: []string{},
			Healthcheck:  &healthcheck,
			PostInstall:  []string{},
		},
		BaseImage: "",
		Version:   "7.4",
		Infer:     infer,
		Stages: map[string]DerivedStage{
			"dev": {
				DeriveFrom: "base",
				Dev:        &isDev,
			},
			"prod": {
				DeriveFrom: "base",
				Dev:        &isNotDev,
			},
		},
	}
}

func NewKind(genericDef *builddef.BuildDef) (Definition, error) {
	def := defaultDefinition()

	decoderConf := mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		Result:           &def,
	}
	decoder, err := mapstructure.NewDecoder(&decoderConf)
	if err != nil {
		return def, err
	}

	if err := decoder.Decode(genericDef.RawConfig); err != nil {
		err := xerrors.Errorf("could not decode build manifest: %w", err)
		return def, err
	}

	if err := yaml.Unmarshal(genericDef.RawLocks, &def.Locks); err != nil {
		err := xerrors.Errorf("could not decode lock manifest: %w", err)
		return def, err
	}

	if def.Webserver != nil {
		*def.Webserver = webserver.DefaultDefinition().Merge(*def.Webserver)
	}

	def.MajMinVersion = extractMajMinVersion(def.Version)

	if def.BaseImage == "" {
		baseImages, ok := defaultBaseImages[def.MajMinVersion]
		if !ok {
			return def, xerrors.Errorf("no default base image defined for PHP v%s, you have to define it by yourself in your zbuild file", def.MajMinVersion)
		}
		if *def.BaseStage.FPM {
			def.BaseImage = baseImages.FPM
		} else {
			def.BaseImage = baseImages.CLI
		}
	}

	return def, nil
}

// Definition holds the specialized config parameters for php images.
// It represents the "base" stage and as such holds the PHP version (ths is the
// only parameter that can't be overriden by derived stages).
type Definition struct {
	BaseStage Stage `mapstructure:",squash"`

	BaseImage     string                  `mapstructure:"base"`
	Version       string                  `mapstructure:"version"`
	MajMinVersion string                  `mapstructure:"-"`
	Infer         bool                    `mapstructure:"infer"`
	Stages        map[string]DerivedStage `mapstructure:"stages"`
	Webserver     *webserver.Definition   `mapstructure:"webserver"`

	Locks DefinitionLocks `mapstructure:"-"`
}

// Stage holds all the properties from the base stage that could also be
// overriden by derived stages.
type Stage struct {
	ExternalFiles     []llbutils.ExternalFile `mapstructure:"external_files"`
	SystemPackages    map[string]string       `mapstructure:"system_packages"`
	FPM               *bool                   `mapstructure:",omitempty"`
	Command           *[]string               `mapstructure:"command"`
	Extensions        map[string]string       `mapstructure:"extensions"`
	GlobalDeps        map[string]string       `mapstructure:"global_deps"`
	ConfigFiles       PHPConfigFiles          `mapstructure:"config_files"`
	ComposerDumpFlags *ComposerDumpFlags      `mapstructure:"composer_dump"`
	Sources           []string                `mapstructure:"sources"`
	Integrations      []string                `mapstructure:"integrations"`
	StatefulDirs      []string                `mapstructure:"stateful_dirs"`
	Healthcheck       *bool                   `mapstructure:"healthcheck"`
	PostInstall       []string                `mapstructure:"post_install"`
}

func (s Stage) Copy() Stage {
	new := Stage{
		ExternalFiles:     make([]llbutils.ExternalFile, len(s.ExternalFiles)),
		SystemPackages:    map[string]string{},
		FPM:               s.FPM,
		Command:           s.Command,
		Extensions:        map[string]string{},
		GlobalDeps:        map[string]string{},
		ConfigFiles:       s.ConfigFiles.Copy(),
		ComposerDumpFlags: s.ComposerDumpFlags,
		Sources:           make([]string, len(s.Sources)),
		Integrations:      make([]string, len(s.Integrations)),
		StatefulDirs:      make([]string, len(s.StatefulDirs)),
		Healthcheck:       s.Healthcheck,
		PostInstall:       make([]string, len(s.PostInstall)),
	}

	copy(new.ExternalFiles, s.ExternalFiles)
	copy(new.Sources, s.Sources)
	copy(new.Integrations, s.Integrations)
	copy(new.StatefulDirs, s.StatefulDirs)
	copy(new.PostInstall, s.PostInstall)

	for name, constraint := range s.SystemPackages {
		new.SystemPackages[name] = constraint
	}
	for name, constraint := range s.Extensions {
		new.Extensions[name] = constraint
	}
	for name, constraint := range s.GlobalDeps {
		new.GlobalDeps[name] = constraint
	}

	return new
}

func (s Stage) Merge(overriding Stage) Stage {
	new := s.Copy()
	new.ExternalFiles = append(new.ExternalFiles,
		overriding.ExternalFiles...)
	new.ConfigFiles = new.ConfigFiles.Merge(overriding.ConfigFiles)
	new.Sources = append(new.Sources, overriding.Sources...)
	new.Integrations = append(new.Integrations, overriding.Integrations...)
	new.StatefulDirs = append(new.StatefulDirs, overriding.StatefulDirs...)
	new.PostInstall = append(new.PostInstall, overriding.PostInstall...)

	for name, constraint := range overriding.SystemPackages {
		new.SystemPackages[name] = constraint
	}
	if overriding.FPM != nil {
		fpm := *overriding.FPM
		new.FPM = &fpm
	}
	if overriding.Command != nil {
		cmd := *overriding.Command
		new.Command = &cmd
	}
	for name, constraint := range overriding.Extensions {
		new.Extensions[name] = constraint
	}
	for name, constraint := range overriding.GlobalDeps {
		new.GlobalDeps[name] = constraint
	}
	if overriding.ComposerDumpFlags != nil {
		dumpFlags := *overriding.ComposerDumpFlags
		new.ComposerDumpFlags = &dumpFlags
	}
	if overriding.Healthcheck != nil {
		healthcheck := *overriding.Healthcheck
		new.Healthcheck = &healthcheck
	}

	return new
}

type DerivedStage struct {
	Stage `mapstructure:",squash"`

	DeriveFrom string `mapstructure:"derive_from"`
	// Dev marks if this is a dev stage (with lighter build process). It's used
	// as a pointer to distinguish when the value is nil or when it's false. In
	// the former case, the value from the parent stage is used.
	Dev *bool `mapstructure:"dev"`
}

type PHPConfigFiles struct {
	IniFile       *string `mapstructure:"php.ini"`
	FPMConfigFile *string `mapstructure:"fpm.conf"`
}

func (base PHPConfigFiles) Copy() PHPConfigFiles {
	new := PHPConfigFiles{}

	if base.IniFile != nil {
		iniFile := *base.IniFile
		new.IniFile = &iniFile
	}
	if base.FPMConfigFile != nil {
		fpmConfigFile := *base.FPMConfigFile
		new.FPMConfigFile = &fpmConfigFile
	}

	return new
}

func (base PHPConfigFiles) Merge(overriding PHPConfigFiles) PHPConfigFiles {
	new := base.Copy()

	if overriding.IniFile != nil {
		iniFile := *overriding.IniFile
		new.IniFile = &iniFile
	}
	if overriding.FPMConfigFile != nil {
		fpmConfigFile := *overriding.FPMConfigFile
		new.FPMConfigFile = &fpmConfigFile
	}

	return new
}

// ComposerDumpFlags represents the optimization flags taken by Composer for
// `composer dump-autoloader`. Only advanced optimizations can be enabled, as
// the --optimize flag is automatically added whenever building images, except
// for dev stage (see cacheWarmup()).
type ComposerDumpFlags struct {
	// APCU enables --apcu flag during composer dump (will use APCu to cache found/not found classes)
	APCU bool `mapstructure:"apcu"`
	// ClassmapAuthoritative enables the matching optimization flag during composer dump.
	ClassmapAuthoritative bool `mapstructure:"classmap_authoritative"`
}

func (fl ComposerDumpFlags) Flags() (string, error) {
	if fl.APCU && fl.ClassmapAuthoritative {
		return "", xerrors.New("you can't use both --apcu and --classmap-authoritative flags. See https://getcomposer.org/doc/articles/autoloader-optimization.md")
	}

	flags := "--no-dev --optimize"
	if fl.APCU {
		flags += " --apcu"
	}
	if fl.ClassmapAuthoritative {
		flags += " --classmap-authoritative"
	}
	return flags, nil
}

// StageDefinition represents the config of stage once it got merged with all
// its ancestors.
type StageDefinition struct {
	Stage
	Name           string
	BaseImage      string
	Version        string
	MajMinVersion  string
	Infer          bool
	Dev            bool
	LockedPackages map[string]string
	PlatformReqs   map[string]string
	Webserver      *webserver.Definition
	Locks          StageLocks
}

func (def *Definition) ResolveStageDefinition(
	stageName string,
	composerLockLoader func(*StageDefinition) error,
	withLocks bool,
) (StageDefinition, error) {
	var stageDef StageDefinition
	stages, err := def.resolveStageChain(stageName)
	if err != nil {
		return stageDef, err
	}

	stageDef = mergeStages(def, stages...)
	stageDef.Name = stageName

	// @TODO: this should not be called here as composer.lock content
	// won't change between stage resolution
	if err := composerLockLoader(&stageDef); err != nil {
		return stageDef, err
	}

	if *stageDef.FPM == false && stageDef.Command == nil {
		return stageDef, xerrors.New("FPM mode is disabled but no command was provided")
	}

	if err := addIntegrations(&stageDef); err != nil {
		return stageDef, err
	}

	if !def.Infer {
		return stageDef, nil
	}

	if stageDef.Dev == false && *stageDef.FPM == true {
		stageDef.Extensions["apcu"] = "*"
		stageDef.Extensions["opcache"] = "*"
	}
	for name, constraint := range stageDef.PlatformReqs {
		if _, ok := stageDef.Extensions[name]; !ok {
			stageDef.Extensions[name] = constraint
		}
	}

	inferExtensions(&stageDef)
	inferSystemPackages(&stageDef)

	if !withLocks {
		return stageDef, nil
	}

	locks, ok := def.Locks.Stages[stageName]
	if !ok {
		return stageDef, xerrors.Errorf(
			"no locks available for stage %q. Please update your lockfile", stageName)
	}
	stageDef.Locks = locks

	return stageDef, nil
}

func (def *Definition) resolveStageChain(name string) ([]DerivedStage, error) {
	stages := make([]DerivedStage, 0, len(def.Stages))
	resolvedStages := map[string]struct{}{}
	current := name

	for current != "" && current != "base" {
		if _, ok := resolvedStages[current]; ok {
			return stages, xerrors.Errorf(
				"there's a cyclic dependency between %q and itself", current)
		}

		stage, ok := def.Stages[current]
		if !ok {
			return stages, xerrors.Errorf("stage %q not found", current)
		}

		stages = append(stages, stage)
		resolvedStages[current] = struct{}{}
		current = stage.DeriveFrom
	}

	return stages, nil
}

func extractMajMinVersion(versionString string) string {
	segments := strings.SplitN(versionString, ".", 3)
	return fmt.Sprintf("%s.%s", segments[0], segments[1])
}

func mergeStages(base *Definition, stages ...DerivedStage) StageDefinition {
	stageDef := StageDefinition{
		BaseImage:      base.BaseImage,
		Version:        base.Version,
		MajMinVersion:  base.MajMinVersion,
		Infer:          base.Infer,
		Stage:          base.BaseStage.Copy(),
		PlatformReqs:   map[string]string{},
		LockedPackages: map[string]string{},
		Webserver:      base.Webserver,
	}

	stages = reverseStages(stages)
	for _, stage := range stages {
		stageDef.Stage = stageDef.Stage.Merge(stage.Stage)

		if stage.Dev != nil {
			stageDef.Dev = *stage.Dev
		}
	}

	if *stageDef.FPM == false {
		stageDef.ConfigFiles.FPMConfigFile = nil
	}
	if stageDef.Dev || *stageDef.FPM {
		healthcheck := false
		stageDef.Healthcheck = &healthcheck
	}

	return stageDef
}

func reverseStages(stages []DerivedStage) []DerivedStage {
	reversed := make([]DerivedStage, len(stages))
	i := 1

	for _, stage := range stages {
		id := len(stages) - i
		reversed[id] = stage
		i++

	}

	return reversed
}

var phpExtDirs = map[string]string{
	"7.2": "/usr/local/lib/php/extensions/no-debug-non-zts-20170718/",
	"7.3": "/usr/local/lib/php/extensions/no-debug-non-zts-20180731/",
	"7.4": "/usr/local/lib/php/extensions/no-debug-non-zts-20190902/",
}

func addIntegrations(stageDef *StageDefinition) error {
	for _, integration := range stageDef.Integrations {
		switch integration {
		case "blackfire":
			dest := path.Join(phpExtDirs[stageDef.MajMinVersion], "blackfire.so")
			stageDef.ExternalFiles = append(stageDef.ExternalFiles, llbutils.ExternalFile{
				URL:         "https://blackfire.io/api/v1/releases/probe/php/linux/amd64/72",
				Compressed:  true,
				Pattern:     "blackfire-*.so",
				Destination: dest,
				Mode:        0644,
			})
		}
	}

	if *stageDef.Healthcheck {
		stageDef.ExternalFiles = append(stageDef.ExternalFiles, llbutils.ExternalFile{
			URL:         "https://github.com/NiR-/fcgi-client/releases/download/v0.1.0/fcgi-client.phar",
			Destination: "/usr/local/bin/fcgi-client",
			Mode:        0750,
			Owner:       "1000:1000",
		})
	}

	return nil
}

// List of extensions preinstalled in official PHP images. Fortunately enough,
// currently all images have the same set of preinstalled extensions.
//
// This list has been obtained using:
// docker run --rm -t php:7.2-fpm-buster php -r 'var_dump(get_loaded_extensions());'
var preinstalledExtensions = []string{
	"core",
	"ctype",
	"curl",
	"date",
	"dom",
	"fileinfo",
	"filter",
	"ftp",
	"hash",
	"iconv",
	"json",
	"libxml",
	"mbstring",
	"mysqlnd",
	"openssl",
	"pcre",
	"pdo",
	"pdo_sqlite",
	"phar",
	"posix",
	"readline",
	"reflection",
	"session",
	"simplexml",
	"sodium",
	"spl",
	"sqlite3",
	"standard",
	"tokenizer",
	"xml",
	"xmlreader",
	"xmlwriter",
	"zlib",
}

func inferExtensions(def *StageDefinition) {
	// soap extension needs sockets extension to work properly
	if _, ok := def.Extensions["soap"]; ok {
		if _, ok := def.Extensions["sockets"]; !ok {
			def.Extensions["sockets"] = "*"
		}
	}

	// Add zip extension if it's missing as it's used by composer to install packages.
	if _, ok := def.Extensions["zip"]; !ok {
		def.Extensions["zip"] = "*"
	}
}

func inferSystemPackages(def *StageDefinition) {
	systemPackages := map[string]string{
		"libpcre3-dev": "*",
	}

	for ext := range def.Extensions {
		deps, ok := extensionsDeps[ext]
		if !ok {
			continue
		}

		for name, ver := range deps {
			systemPackages[name] = ver
		}
	}

	// Add unzip and git packages as they're used by Composer
	if _, ok := def.SystemPackages["unzip"]; !ok {
		systemPackages["unzip"] = "*"
	}
	if _, ok := def.SystemPackages["git"]; !ok {
		systemPackages["git"] = "*"
	}

	for name, constraint := range systemPackages {
		if _, ok := def.SystemPackages[name]; !ok {
			def.SystemPackages[name] = constraint
		}
	}
}
