package concept

import (
	"errors"
	"fmt"
	"github.com/dityaaa/concept/database"
	"github.com/dityaaa/concept/internal/natsort"
	"github.com/dityaaa/concept/source"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Concept struct {
	databaseDriver database.Driver
	sourceDriver   source.Driver

	versions   []string
	migrations map[string]*Migration

	pattern *regexp.Regexp

	latestErr    error
	unpairedRevs int
	outOfOrder   bool

	latestSourceVersion   string
	latestDatabaseVersion string

	hooks *Hooks
}

func New(databaseUrl, sourceUrl string) (*Concept, error) {
	dbDrv, err := database.Open(databaseUrl)
	if err != nil {
		return nil, err
	}

	scDrv, err := source.Open(sourceUrl)
	if err != nil {
		return nil, err
	}

	return NewWithInstance(dbDrv, scDrv)
}

func NewWithInstance(database database.Driver, source source.Driver) (*Concept, error) {
	if database == nil && source == nil {
		return nil, errors.New("concept: database and source instance cannot be nil")
	}

	if database == nil {
		return nil, errors.New("concept: database instance cannot be nil")
	}

	if source == nil {
		return nil, errors.New("concept: source instance cannot be nil")
	}

	inst := &Concept{
		databaseDriver: database,
		sourceDriver:   source,
		versions:       make([]string, 0),
		migrations:     make(map[string]*Migration, 0),
		pattern:        regexp.MustCompile(`(\d+?)(?:_(\w*))?(?:\.(adv|rev))?.sql$`),
	}
	inst.ClearHooks()

	return inst, nil
}

func (i *Concept) SetHooks(hooks *Hooks) {
	types := reflect.TypeOf(Hooks{}).Elem()

	for c := 0; c < types.NumField(); c++ {
		srcField := reflect.ValueOf(hooks).Elem().Field(c)
		destField := reflect.ValueOf(i.hooks).Elem().Field(c)
		if !srcField.IsNil() {
			destField.Set(srcField)
		}
	}
}

func (i *Concept) ClearHooks() {
	i.hooks = &Hooks{
		PreMigrate:   func(m *Migration) {},
		PostMigrate:  func(m *Migration) {},
		MigrateErr:   func(m *Migration, err error) {},
		PreRollback:  func(m *Migration) {},
		PostRollback: func(m *Migration) {},
		RollbackErr:  func(m *Migration, err error) {},
	}
}

func (i *Concept) Create(name string, rev bool) ([]string, error) {
	latestVer, err := strconv.Atoi(i.latestSourceVersion)
	if err != nil {
		return nil, errors.New("concept: create only support sequential version name")
	}

	latestDatabaseVersion, err := strconv.Atoi(i.latestDatabaseVersion)
	if err != nil {
		return nil, errors.New("concept: create only support sequential version name")
	}

	if latestVer < latestDatabaseVersion {
		return nil, errors.New("concept: outdated source migration")
	}

	latestVer++
	name = fmt.Sprintf("%05d_%s", latestVer, name)
	// TODO: support for customizable adv/rev suffix
	files := []string{
		name + ".sql",
	}
	if rev == true {
		files = []string{
			name + ".adv.sql",
			name + ".rev.sql",
		}
	}

	created := false
	for _, filename := range files {
		if err = i.sourceDriver.Touch(filename); err != nil {
			break
		}
		created = true
	}

	if err != nil {
		if created {
			err = i.sourceDriver.Remove(files[0])
			if err != nil {
				return nil, err
			}
		}
		return nil, err
	}

	return files, nil
}

func (i *Concept) Migrate(steps int) error {
	count := 0
	for _, version := range i.versions {
		mg := i.migrations[version]
		if (mg.State&pendingState) != pendingState && (mg.State&undoneState) != undoneState {
			continue
		}

		count++
		if steps < 0 && count > steps {
			break
		}

		if mg.State&failedState > 0 {
			return fmt.Errorf("last database migration is failed. manual cleaning needed at version: %s", mg.Version)
		}

		hs := &database.History{
			Mode:        AdvanceDirection,
			Version:     mg.Version,
			ScriptName:  mg.AdvanceScript.Identifier,
			Description: mg.Description,
			Checksum:    mg.AdvanceScript.Checksum(),
			AppliedAt:   uint64(time.Now().Unix()),
		}
		err := i.databaseDriver.Write(hs)

		currentTime := time.Now()
		executionTime := uint32(currentTime.UnixMilli())
		if err := i.databaseDriver.Run(mg.AdvanceScript); err != nil {
			return err
		}
		executionTime = uint32(time.Now().UnixMilli()) - executionTime

		mg.AppliedAt = hs.AppliedAt
		mg.ExecutionTime = executionTime
		hs.ExecutionTime = executionTime
		hs.Success = true
		mg.State &= ^pendingState
		mg.State &= ^undoneState

		err = i.databaseDriver.Write(hs)
		if err != nil {
			mg.State |= failedState
			return err
		}

		mg.State |= successState
	}

	return nil
}

func (i *Concept) Rollback(steps int) error {
	count := 0
	for c := len(i.versions) - 1; c >= 0 && (steps < 0 || count < steps); c-- {
		version := i.versions[c]
		mg := i.migrations[version]

		if (mg.State&availableState) != availableState || (mg.State&undoneState) == undoneState {
			continue
		}

		count++

		hs := &database.History{
			Mode:        ReverseDirection,
			Version:     mg.Version,
			ScriptName:  mg.ReverseScript.Identifier,
			Description: mg.Description,
			Checksum:    mg.ReverseScript.Checksum(),
			AppliedAt:   uint64(time.Now().Unix()),
		}
		err := i.databaseDriver.Write(hs)

		executionTime := uint32(hs.AppliedAt)

		if err := i.databaseDriver.Run(mg.ReverseScript); err != nil {
			return err
		}

		executionTime = uint32(time.Now().UnixMilli()) - executionTime
		mg.ExecutionTime = executionTime
		hs.ExecutionTime = executionTime
		hs.Success = true
		mg.State |= pendingState

		err = i.databaseDriver.Write(hs)
		if err != nil {
			mg.State |= failedState
			return err
		}

		mg.State |= undoneState
	}

	return nil
}

func (i *Concept) Refresh() error {
	return i.rebuild()
}

func (i *Concept) Get() ([]*Migration, error) {
	if i.latestErr != nil {
		return nil, i.latestErr
	}

	migrations := make([]*Migration, 0, len(i.migrations))
	for _, version := range i.versions {
		migrations = append(migrations, i.migrations[version])
	}
	return migrations, nil
}

func (i *Concept) Purge() error {
	if errs := i.databaseDriver.Purge(); len(errs) > 0 {
		return fmt.Errorf("concept: purge completed with %v errors", len(errs))
	}

	return nil
}

func (i *Concept) rebuild() error {
	for i.sourceDriver.Next() {
		mg, err := i.sourceDriver.Read()
		if err != nil {
			i.latestErr = err
			return err
		}
		i.latestErr = i.appendSource(mg)
		if i.latestErr != nil {
			return i.latestErr
		}
	}

	if i.unpairedRevs > 0 {
		unpairedRevs := make([]string, 0, i.unpairedRevs)
		for _, version := range i.versions {
			migration := i.migrations[version]
			if migration.AdvanceScript == nil {
				unpairedRevs = append(unpairedRevs, migration.ReverseScript.Identifier)
			}
		}
		natsort.Sort(unpairedRevs)
		return fmt.Errorf("concept: found %v rev without adv migration %v", i.unpairedRevs, unpairedRevs)
	}

	var histories []*database.History
	histories, i.latestErr = i.databaseDriver.Read()
	if i.latestErr != nil {
		return i.latestErr
	}

	for _, history := range histories {
		// TODO: make databaseAppend independent from appendSource, so we can sync while appending from source
		i.latestErr = i.databaseAppend(history)
		if i.latestErr != nil {
			return i.latestErr
		}
	}

	natsort.Sort(i.versions)

	unavailable := false
	for c := len(i.versions) - 1; c >= 0; c-- {
		version := i.versions[c]
		migration := i.migrations[version]

		if (migration.State & availableState) == 0 {
			unavailable = true
		}

		if unavailable {
			migration.State &^= availableState
		}
	}

	return nil
}

func (i *Concept) databaseAppend(history *database.History) error {
	item, exists := i.migrations[history.Version]
	if !exists {
		// we skip reverse migration
		if Direction(history.Mode) == ReverseDirection {
			return nil
		}

		item = &Migration{
			Version:       history.Version,
			Description:   history.Description,
			AppliedBy:     history.AppliedBy,
			AppliedAt:     history.AppliedAt,
			ExecutionTime: history.ExecutionTime,
			State:         failedState | missingState,
		}

		if history.Success {
			item.State = successState | missingState
		}

		i.migrations[item.Version] = item
		i.versions = append(i.versions, item.Version)

		return nil
	}

	if item.Description != history.Description {
		return fmt.Errorf("concept: possibly wrong migration/database (mismatch migration description)")
	}

	item.State = failedState
	if history.Success {
		item.State = successState
	}

	// migration history only save checksum for the advance script. so here, we
	// only check for the advance script.
	checksumAvailable := item.AdvanceScript != nil && history.Checksum != "" && item.AdvanceScript.Checksum() != ""
	checksumMismatch := checksumAvailable && item.AdvanceScript.Checksum() != history.Checksum
	if checksumMismatch {
		item.State |= futureState
	}

	if item.ReverseScript != nil {
		item.State |= availableState
	}

	return nil
}

func (i *Concept) appendSource(migration *source.Migration) error {
	script, err := i.parse(migration.Identifier)
	if err != nil {
		return err
	}

	item, exists := i.migrations[script.Version]
	if !exists {
		item = &Migration{
			Version:     script.Version,
			Description: script.Description,
			State:       pendingState,
		}

		item.AdvanceScript = script
		if script.Direction == ReverseDirection {
			i.unpairedRevs++
			item.State = unknownState
			item.AdvanceScript = nil
			item.ReverseScript = script
		}

		i.migrations[script.Version] = item
		i.versions = append(i.versions, script.Version)
		return nil
	}

	duplicateAdvanceMigration := script.Direction == AdvanceDirection && item.AdvanceScript != nil
	duplicateReverseMigration := script.Direction == ReverseDirection && item.ReverseScript != nil
	if duplicateAdvanceMigration || duplicateReverseMigration {
		return fmt.Errorf("concept: duplicate migration %v", script.Identifier)
	}

	prevScript := item.AdvanceScript
	if prevScript == nil {
		prevScript = item.ReverseScript
	}

	if prevScript.Description != script.Description {
		return fmt.Errorf("concept: migration script pattern is mismatch [%v ; %v]", prevScript.Identifier, script.Identifier)
	}

	if script.Direction == ReverseDirection {
		item.ReverseScript = script
		return nil
	}

	item.State &^= missingState
	item.AdvanceScript = script

	if item.ReverseScript != nil {
		i.unpairedRevs--
	}

	return nil
}

func (i *Concept) parse(identifier string) (*Script, error) {
	matches := i.pattern.FindStringSubmatch(identifier)
	if matches == nil {
		return nil, errors.New("concept: encounter invalid migration identifier")
	}

	direction := Direction(strings.ToUpper(matches[3]))
	if direction == "" {
		direction = AdvanceDirection
	}

	return &Script{
		Version:     matches[1],
		Identifier:  identifier,
		Description: matches[2],
		Direction:   direction,
	}, nil
}
