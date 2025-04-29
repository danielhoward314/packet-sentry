-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE packet_events (
    id SERIAL,
    event_time TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, event_time),
    os_unique_identifier TEXT NOT NULL,
    bpf TEXT NOT NULL,
    interface VARCHAR(255) NOT NULL,
    promiscuous BOOLEAN,
    snap_length INT,
    capture_length INT NOT NULL,
    original_length INT NOT NULL,
    interface_index INT NOT NULL,
    truncated BOOLEAN NOT NULL,
    ip_version TEXT,
    ip_src TEXT,
    ip_dst TEXT,
    ip_ttl INT,
    ip_hop_limit INT,
    ip_protocol VARCHAR(255),
    tcp_src_port INT,
    tcp_dst_port INT,
    tcp_seq BIGINT,
    tcp_ack BIGINT,
    tcp_fin BOOLEAN,
    tcp_syn BOOLEAN,
    tcp_rst BOOLEAN,
    tcp_psh BOOLEAN,
    tcp_ack_flag BOOLEAN,
    tcp_urg BOOLEAN,
    tcp_window INT,
    udp_src_port INT,
    udp_dst_port INT,
    udp_length INT,
    tls_record_count INT
);

-- Make it a hypertable
SELECT create_hypertable('packet_events', 'event_time');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS packet_events;
-- +goose StatementEnd
