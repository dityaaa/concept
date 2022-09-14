// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migration

import (
	"crypto/md5"
	"fmt"
	"github.com/dityaaa/concept/database"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Hooks struct {
	PreMigrate  func(dt Data)
	PostMigrate func(dt Data)
	MigrateErr  func(dt Data, err error)

	PreRollback  func(dt Data)
	PostRollback func(dt Data)
	RollbackErr  func(dt Data, err error)
}

type Config struct {
	Path            string
	AllowOutOfOrder bool
	Database        database.Database

	*Hooks
}

type Data struct {
	Sequence      *uint64
	Version       string
	ScriptName    string
	Description   string
	Checksum      string
	AppliedBy     string
	AppliedAt     uint64
	ExecutionTime uint32
	Status        State
}

type Properties struct {
	path            string
	db              database.Database
	allowOutOfOrder bool

	localKeys  []string
	localItems map[string]*Data

	dbaseKeys  []string
	dbaseItems map[string]*Data

	keys       []string
	items      map[string]*Data
	outOfOrder bool

	hooks *Hooks
}

func New(config *Config) (*Properties, error) {
	hooks := config.Hooks
	if hooks == nil {
		hooks = &Hooks{}
	}

	return &Properties{
		path:            config.Path,
		db:              config.Database,
		allowOutOfOrder: config.AllowOutOfOrder,
		hooks:           hooks,
	}, nil
}

func (p *Properties) Create(name string, rev bool) ([]string, error) {
	err := p.readLocal()
	if err != nil {
		return nil, err
	}

	version := 1
	if len(p.localKeys) > 0 {
		version, _ = strconv.Atoi(p.localKeys[len(p.localKeys)-1])
		version++
	}

	name = fmt.Sprintf("%05d_%s", version, name)
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

	success := false
	for _, filename := range files {
		if _, err = os.Create(path.Join(p.path, filename)); err != nil {
			break
		}
		success = true
	}

	if err != nil {
		if success {
			_ = os.Remove(path.Join(p.path, files[0]))
		}
		return nil, err
	}

	return files, nil
}

func (p *Properties) Migrate() error {
	err := p.sync()
	if err != nil {
		return err
	}

	for _, key := range p.keys {
		mg := p.items[key]
		if (mg.Status&PendingState) != PendingState && (mg.Status&UndoneState) != UndoneState {
			continue
		}

		if strings.HasSuffix(mg.ScriptName, ".rev.sql") {
			mg.ScriptName = strings.TrimSuffix(mg.ScriptName, ".rev.sql")
			mg.ScriptName = mg.ScriptName + ".adv.sql"
		}

		file, err := os.ReadFile(filepath.Join(p.path, mg.ScriptName))
		if err != nil {
			p.hooks.MigrateErr(*mg, err)
			return err
		}

		p.hooks.PreMigrate(*mg)

		row, err := p.db.Insert("ADV", mg.Version, mg.ScriptName, mg.Description, mg.Checksum)
		if err != nil {
			p.hooks.MigrateErr(*mg, err)
			return err
		}

		mg.Sequence = &row.Sequence
		mg.AppliedAt = row.AppliedAt
		executionTime := uint32(row.AppliedAt)

		if err := p.db.Exec(string(file)); err != nil {
			p.hooks.MigrateErr(*mg, err)
			return err
		}
		executionTime = uint32(time.Now().UnixMilli()) - executionTime

		mg.ExecutionTime = executionTime
		mg.Status &= ^PendingState
		mg.Status &= ^UndoneState

		err = p.db.Update(row.Sequence, executionTime, true)
		if err != nil {
			mg.Status |= FailedState
			p.hooks.MigrateErr(*mg, err)
			return err
		}

		mg.Status |= SuccessState
		p.hooks.PostMigrate(*mg)
	}

	return nil
}

func (p *Properties) Rollback(steps int) error {
	err := p.sync()
	if err != nil {
		return err
	}

	counter := 0
	for i := len(p.keys) - 1; i >= 0 && counter < steps; i-- {
		key := p.keys[i]
		mg := &(*p.items[key])
		mg.Sequence = nil
		// TODO: support for customizable adv/rev suffix
		mg.ScriptName = strings.TrimSuffix(mg.ScriptName, "adv.sql") + "rev.sql"

		if (mg.Status&AvailableState) != AvailableState || (mg.Status&UndoneState) == UndoneState {
			continue
		}

		counter++

		file, err := os.ReadFile(filepath.Join(p.path, mg.ScriptName))
		if err != nil {
			if p.hooks.MigrateErr != nil {
				p.hooks.MigrateErr(*mg, err)
			}
			return err
		}

		if p.hooks.PreRollback != nil {
			p.hooks.PreRollback(*mg)
		}

		row, err := p.db.Insert("REV", string(mg.Version), mg.ScriptName, mg.Description, mg.Checksum)
		if err != nil {
			if p.hooks.RollbackErr != nil {
				p.hooks.RollbackErr(*mg, err)
			}
			return err
		}

		mg.Sequence = &row.Sequence
		mg.AppliedAt = row.AppliedAt
		executionTime := uint32(row.AppliedAt)

		if err := p.db.Exec(string(file)); err != nil {
			if p.hooks.RollbackErr != nil {
				p.hooks.RollbackErr(*mg, err)
			}
			return err
		}
		executionTime = uint32(time.Now().UnixMilli()) - executionTime

		mg.ExecutionTime = executionTime
		mg.Status &= ^PendingState

		err = p.db.Update(row.Sequence, executionTime, true)
		if err != nil {
			mg.Status |= FailedState
			p.hooks.RollbackErr(*mg, err)
			return err
		}

		mg.Status |= UndoneState
		if p.hooks.PostRollback != nil {
			p.hooks.PostRollback(*mg)
		}

	}

	return nil
}

func (p *Properties) Get() ([]*Data, error) {
	err := p.sync()
	if err != nil {
		return nil, err
	}

	res := make([]*Data, 0)
	for _, key := range p.keys {
		res = append(res, p.items[key])
	}

	return res, nil
}

func (p *Properties) readLocal() error {
	entries, err := os.ReadDir(p.path)
	if err != nil {
		return err
	}

	p.localKeys = make([]string, 0)
	p.localItems = make(map[string]*Data, 0)
	prevVersion := "0"
	versionPair := 0
	// adv = advance; rev = revert/reverse
	// TODO: support for customizable adv/rev suffix
	filePattern := regexp.MustCompile(`^(\d+?)(_\w*)?(?:\.(adv|rev))?.sql$`)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matches := filePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			return fmt.Errorf("incorrect migration name detected: %s", entry.Name())
		}

		version := matches[1]
		description := strings.TrimPrefix(matches[2], "_")
		mode := matches[3]
		if mode == "" {
			mode = "adv"
		}

		if versionPair > 2 && version == prevVersion {
			return fmt.Errorf("duplicate version encountered: %s", entry.Name())
		}

		if version == prevVersion && mode == "adv" {
			return fmt.Errorf("abnormal migration filename: %s", entry.Name())
		}

		if version != prevVersion && mode == "rev" {
			return fmt.Errorf("reverse migration without advance file: %s", entry.Name())
		}

		versionPair++

		if mode == "rev" {
			versionPair = 0
			p.localItems[version].Status |= AvailableState
			continue
		}

		if mode == "adv" && prevVersion != version {
			versionPair = 0
		}

		p.localKeys = append(p.localKeys, version)

		file, err := os.ReadFile(filepath.Join(p.path, entry.Name()))
		if err != nil {
			return err
		}

		checksum := fmt.Sprintf("%x", md5.Sum(file))

		data := &Data{
			Version:     version,
			ScriptName:  entry.Name(),
			Description: description,
			Checksum:    checksum,
			Status:      UnknownState,
		}

		p.localItems[version] = data
		prevVersion = version
	}

	sortVersion(p.localKeys)

	return nil
}

func (p *Properties) readDbase() error {
	data, err := p.db.Read("*")
	if err != nil {
		return err
	}

	p.dbaseItems = translate(data)
	sortVersion(keys(p.dbaseItems))

	return nil
}

func (p *Properties) sync() error {
	err := p.readLocal()
	if err != nil {
		return err
	}

	err = p.readDbase()
	if err != nil {
		return err
	}

	p.keys = unique(append(p.localKeys, p.dbaseKeys...))
	sortVersion(p.keys)
	p.items = make(map[string]*Data, len(p.keys))

	size := len(p.keys)
	prevAvail := true
	prevState := UnknownState
	p.outOfOrder = false

	// iterating backward to check if backward migration is possible.
	// assuming we have 5 histories, with 1, 4, and 5 being available.
	// only migration number 4 and 5 can be rolled back, because migration
	// number 2 and 3 does not available, so number 1 cannot be reached
	// for backward migration.
	for i := size - 1; i >= 0; i-- {
		key := p.keys[i]
		dbase, dbaseExists := p.dbaseItems[key]
		local, localExists := p.localItems[key]

		// check if migration is out of order. out of order means that
		// pending migration is exists in the middle of applied migrations
		if localExists && !dbaseExists && (prevState&PendingState) == 0 && prevState != UnknownState {
			p.outOfOrder = true
		}

		// mark migration as pending if key does not exist in database
		if !dbaseExists {
			local.Status |= PendingState
			p.items[key] = local
			prevState = local.Status
			continue
		}

		// mark migration as missing if key only exists in database
		if !localExists {
			// we add "missing" state (besides failed/success state from database)
			dbase.Status |= MissingState
			p.items[key] = dbase
			prevState = dbase.Status
			continue
		}

		// assume migration script is replaced if local script name is not equal
		// with migrated script name
		if !strings.HasSuffix(dbase.ScriptName, ".rev.sql") && local.ScriptName != dbase.ScriptName {
			p.outOfOrder = true
		}

		// mark migration as future if checksum is not equal
		if local.Checksum != dbase.Checksum {
			dbase.Status |= FutureState
		}

		// mark migration as failed if key exists in local and database,
		// but success value is false
		if (dbase.Status & FailedState) == FailedState {
			p.localItems[key] = dbase
			p.items[key] = dbase
			prevState = dbase.Status
			continue
		}

		if prevAvail && (local.Status&AvailableState) == 0 {
			prevAvail = false
		}

		// propagate state from local migration
		if prevAvail && (local.Status&AvailableState) > 0 {
			dbase.Status |= local.Status & AvailableState
		}

		// mark migration as success
		p.localItems[key] = dbase
		p.items[key] = dbase
		prevState = dbase.Status
	}

	// TODO: allow out of order migration using command flag
	if p.outOfOrder {
		return fmt.Errorf("out of order migration is detected")
	}

	return nil
}
