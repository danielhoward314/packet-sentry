package postgres

import (
	"database/sql"
	"fmt"

	"github.com/danielhoward314/packet-sentry/dao"
	"github.com/danielhoward314/packet-sentry/dao/postgres/queries"
)

type events struct {
	db *sql.DB
}

// NewEvents returns an instance implementing the Events interface
func NewEvents(db *sql.DB) dao.Events {
	return &events{db: db}
}

func (e *events) Read(deviceID string, start string, end string) ([]*dao.Event, error) {
	if deviceID == "" {
		return nil, fmt.Errorf("empty device id")
	}
	if start == "" {
		return nil, fmt.Errorf("empty start")
	}
	if end == "" {
		return nil, fmt.Errorf("empty end")
	}

	var events []*dao.Event
	rows, rowsErr := e.db.Query(
		queries.EventsSelectByDeviceIdDatetime,
		deviceID,
		start,
		end,
	)
	if rowsErr != nil {
		return nil, rowsErr
	}

	for rows.Next() {
		var event dao.Event

		rowErr := rows.Scan(
			&event.EventTime,
			&event.Bpf,
			&event.OriginalLength,
			&event.IpSrc,
			&event.IpDst,
			&event.TcpSrcPort,
			&event.TcpDstPort,
			&event.IpVersion,
		)

		if rowErr != nil {
			return nil, rowErr
		}

		events = append(events, &event)
	}

	return events, nil
}
