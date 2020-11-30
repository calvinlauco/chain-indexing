package event

import (
	"errors"
	"fmt"

	"github.com/crypto-com/chain-indexing/appinterface/rdb"
	entity_event "github.com/crypto-com/chain-indexing/entity/event"
)

const DEFAULT_TABLE = "events"

// block heights per partition table
const PARTITION_SIZE = 5000

// Events table should have the following schema
// | Field   | Data Type | Constraint  |
// | ------- | --------- | ----------- |
// | id      | VARCHAR   | PRIMARY KEY |
// | height  | INT64     | NOT NULL    |
// | name    | VARCHAR   | NOT NULL    |
// | version | INT64     | NOT NULL    |
// | payload | JSONB     | NOT NULL    |

var _ entity_event.Store = &RDbStore{}

// EventStore implemented using relational database
type RDbStore struct {
	rdbHandle *rdb.Handle
	Registry  *entity_event.Registry

	table string
}

func NewRDbStore(handle *rdb.Handle, registry *entity_event.Registry) *RDbStore {
	return &RDbStore{
		rdbHandle: handle,
		Registry:  registry,

		table: DEFAULT_TABLE,
	}
}

// GetLatestHeight returns latest event height, nil if no event is stored
func (store *RDbStore) GetLatestHeight() (*int64, error) {
	sql, args, err := store.rdbHandle.StmtBuilder.Select(
		"MAX(height)",
	).From(
		store.table,
	).ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building latest event height selection SQL: %v", err)
	}

	var latestEventHeight *int64
	if err := store.rdbHandle.QueryRow(sql, args...).Scan(&latestEventHeight); err != nil {
		if errors.Is(err, rdb.ErrNoRows) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("error executing latest event height selection SQL: %v", err)
		}
	}

	return latestEventHeight, nil
}

func (store *RDbStore) GetAllByHeight(height int64) ([]entity_event.Event, error) {
	sql, args, err := store.rdbHandle.StmtBuilder.Select(
		"uuid", "height", "name", "version", "payload",
	).From(
		store.table,
	).Where(
		"height = ?", height,
	).OrderBy("id").ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building get all events by height selection SQL: %v", err)
	}

	var events = make([]entity_event.Event, 0)
	rows, err := store.rdbHandle.Query(sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing get all events by height selection SQL: %v", err)
	}
	for rows.Next() {
		var (
			uuid    string
			height  int64
			name    string
			version int
			payload string
		)

		if err := rows.Scan(&uuid, &height, &name, &version, &payload); err != nil {
			if errors.Is(err, rdb.ErrNoRows) {
				return nil, nil
			} else {
				return nil, fmt.Errorf("error executing get each event by height selection SQL: %v", err)
			}
		}

		event, err := store.Registry.DecodeByType(name, version, []byte(payload))
		if err != nil {
			return nil, fmt.Errorf("error decoding the event string into type: %v", err)
		}

		events = append(events, event)
	}

	return events, nil
}

func (store *RDbStore) Insert(event entity_event.Event) error {
	encodedEvent, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("error encoding event to json: %v", err)
	}
	sql, args, err := store.rdbHandle.StmtBuilder.Insert(
		store.table,
	).Columns(
		"uuid", "height", "name", "version", "payload",
	).Values(
		event.UUID(),
		event.Height(),
		event.Name(),
		event.Version(),
		encodedEvent,
	).ToSql()
	if err != nil {
		return fmt.Errorf("error building event insertion SQL: %v", err)
	}

	execResult, err := store.rdbHandle.Exec(sql, args...)
	if err != nil {
		return fmt.Errorf("error exectuing event insertion SQL: %v", err)
	}
	if execResult.RowsAffected() == 0 {
		return errors.New("error executing event insertion SQL: no rows inserted")
	}

	return nil
}

// InsertAll insert all events into store. It will rollback when the insert fails at any point.
func (store *RDbStore) InsertAll(events []entity_event.Event) error {
	if len(events) == 0 {
		return nil
	}

	stmtBuilder := store.rdbHandle.StmtBuilder.Insert(
		store.table,
	).Columns(
		"uuid", "height", "name", "version", "payload",
	)
	for _, event := range events {
		encodedEvent, err := event.ToJSON()
		if err != nil {
			return fmt.Errorf("error encoding event to json: %v", err)
		}
		stmtBuilder = stmtBuilder.Values(
			event.UUID(),
			event.Height(),
			event.Name(),
			event.Version(),
			encodedEvent,
		)
	}
	sql, args, err := stmtBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("error building event insertion SQL: %v", err)
	}

	execResult, err := store.rdbHandle.Exec(sql, args...)
	if err != nil {
		return fmt.Errorf("error exectuing event insertion SQL: %v", err)
	}
	if execResult.RowsAffected() != int64(len(events)) {
		return errors.New("error executing event insertion SQL: mismatched number of rows inserted")
	}

	return nil
}

// EnsurePartitionTableExists creates events partition table for height, assuming caller calls with increasing height argument.
func (store *RDbStore) EnsurePartitionTableExists(height int64) error {
	if height%PARTITION_SIZE == 0 {
		idx := height / PARTITION_SIZE
		_, err := store.rdbHandle.Exec(fmt.Sprintf("create table if not exists events_%d partition of events for values from (%d) to (%d)", idx, idx*PARTITION_SIZE, (idx+1)*PARTITION_SIZE))
		return err
	}
	return nil
}
