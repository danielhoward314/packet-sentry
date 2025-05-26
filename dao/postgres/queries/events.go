package queries

const EventsSelectByDeviceIdDatetime = `
SELECT
	event_time, bpf, original_length, ip_src,
	ip_dst, tcp_src_port, tcp_dst_port, ip_version
FROM packet_events
WHERE os_unique_identifier = $1
AND event_time BETWEEN $2 AND $3
`
