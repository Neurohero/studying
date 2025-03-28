package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/lib/pq"
)

const (
	host     = "10.10.10.10"
	port     = 1111
	user     = "user"
	password = "password"
	dbname   = "db"
)

var (
	logicalReplicationLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "logical_replication_lag",
			Help: "WAL lag for logical replication",
		},
		[]string{"slot_name", "instance"},
	)

	logicalReplicationFlushLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "logical_replication_flush_lag",
			Help: "WAL flush lag for logical replication",
		},
		[]string{"slot_name", "instance"},
	)

	logicalReplicationIsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "logical_replication_is_active",
			Help: "Whether the logical replication is active",
		},
		[]string{"slot_name", "instance"},
	)
)

func init() {
	prometheus.MustRegister(logicalReplicationLag)
	prometheus.MustRegister(logicalReplicationFlushLag)
	prometheus.MustRegister(logicalReplicationIsActive)
}

func connectToDB() (*sql.DB, error) {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", user, password, host, port, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func collectMetrics(db *sql.DB) error {
	var lag float64
	err := db.QueryRow(`
		SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)
		FROM pg_catalog.pg_replication_slots
		WHERE slot_type = 'logical'
		AND slot_name = 'from_dc1_to_dc2_schema_journaldc1_sub';`).Scan(&lag)
	if err != nil {
		return err
	}
	logicalReplicationLag.WithLabelValues("from_dc1_to_dc2_schema_journaldc1_sub", "dc1-nsud-db:9188").Set(lag)

	// Сбор данных о лаге флаша
	var flushLag float64
	err = db.QueryRow(`
		SELECT pg_wal_lsn_diff(pg_current_wal_lsn(), confirmed_flush_lsn)
		FROM pg_catalog.pg_replication_slots
		WHERE slot_type = 'logical'
		AND slot_name = 'from_dc1_to_dc2_schema_journaldc1_sub';`).Scan(&flushLag)
	if err != nil {
		return err
	}
	logicalReplicationFlushLag.WithLabelValues("from_dc1_to_dc2_schema_journaldc1_sub", "dc1-nsud-db:9188").Set(flushLag)

	var isActive bool
	err = db.QueryRow(`
		SELECT CASE WHEN active THEN 1 ELSE 0 END
		FROM pg_catalog.pg_replication_slots
		WHERE slot_type = 'logical'
		AND slot_name = 'from_dc1_to_dc2_schema_journaldc1_sub';`).Scan(&isActive)
	if err != nil {
		return err
	}
	var activeValue float64
	if isActive {
		activeValue = 1
	} else {
		activeValue = 0
	}
	logicalReplicationIsActive.WithLabelValues("from_dc1_to_dc2_schema_journaldc1_sub", "dc1-nsud-db:9188").Set(activeValue)

	return nil
}

func main() {
	db, err := connectToDB()
	if err != nil {
		os.Exit(1)
	}
	defer db.Close()

	go func() {
		for {
			if err := collectMetrics(db); err != nil {
				// Ошибки не логируются
				os.Exit(1)
			}
			time.Sleep(60 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(":9188", nil); err != nil {
		os.Exit(1)
	}
}

