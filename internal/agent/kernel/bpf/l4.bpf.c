/* L4 connection accounting for the Argus agent.
 *
 * Counts TCP connection establishment and retransmissions per address pair.
 * Aggregation happens in the kernel map so userspace reads a small summary
 * instead of one event per packet.
 *
 * Both programs are stable tracepoints read through CO-RE, so a single
 * little-endian object works on amd64 and arm64.
 */
#include <linux/bpf.h>
#include "kernel_defs.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>

char LICENSE[] SEC("license") = "GPL";

#define AF_INET 2
#define AF_INET6 10
#define IPPROTO_TCP 6

#define TCP_SYN_SENT 2
#define TCP_SYN_RECV 3
#define TCP_LISTEN 10

#define FLOW_OUT 0
#define FLOW_IN 1
#define FLOW_RETRANSMIT 2

#define STAT_EVENTS 0
#define STAT_SKIPPED 1
#define STAT_MAP_ERRORS 2
#define STAT_SLOTS 3

/* The service port is sport on inbound flows and dport on outbound ones, so
 * both are part of the key and userspace picks the relevant side.
 */
struct flow_key {
	__u8 saddr[16];
	__u8 daddr[16];
	__u16 sport;
	__u16 dport;
	__u8 family;
	__u8 direction;
};

struct flow_value {
	__u64 count;
	__u64 last_ns;
};

struct {
	__uint(type, BPF_MAP_TYPE_LRU_HASH);
	__uint(max_entries, 16384);
	__type(key, struct flow_key);
	__type(value, struct flow_value);
} flows SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_ARRAY);
	__uint(max_entries, STAT_SLOTS);
	__type(key, __u32);
	__type(value, __u64);
} stats SEC(".maps");

static __always_inline void bump(__u32 slot)
{
	__u64 *slot_value = bpf_map_lookup_elem(&stats, &slot);
	if (slot_value)
		__sync_fetch_and_add(slot_value, 1);
}

static __always_inline void account(struct flow_key *key)
{
	struct flow_value *existing = bpf_map_lookup_elem(&flows, key);
	if (existing) {
		__sync_fetch_and_add(&existing->count, 1);
		existing->last_ns = bpf_ktime_get_ns();
	} else {
		struct flow_value fresh = {.count = 1, .last_ns = bpf_ktime_get_ns()};
		if (bpf_map_update_elem(&flows, key, &fresh, BPF_ANY)) {
			bump(STAT_MAP_ERRORS);
			return;
		}
	}
	bump(STAT_EVENTS);
}

SEC("tracepoint/sock/inet_sock_set_state")
int argus_inet_sock_set_state(struct trace_event_raw_inet_sock_set_state *ctx)
{
	if (BPF_CORE_READ(ctx, protocol) != IPPROTO_TCP)
		return 0;

	int oldstate = BPF_CORE_READ(ctx, oldstate);
	int newstate = BPF_CORE_READ(ctx, newstate);
	struct flow_key key = {};

	if (newstate == TCP_SYN_SENT)
		key.direction = FLOW_OUT;
	else if (oldstate == TCP_LISTEN && newstate == TCP_SYN_RECV)
		key.direction = FLOW_IN;
	else
		return 0;

	__u16 family = BPF_CORE_READ(ctx, family);
	if (family == AF_INET) {
		key.family = 4;
		BPF_CORE_READ_INTO(&key.saddr, ctx, saddr);
		BPF_CORE_READ_INTO(&key.daddr, ctx, daddr);
	} else if (family == AF_INET6) {
		key.family = 6;
		BPF_CORE_READ_INTO(&key.saddr, ctx, saddr_v6);
		BPF_CORE_READ_INTO(&key.daddr, ctx, daddr_v6);
	} else {
		bump(STAT_SKIPPED);
		return 0;
	}

	key.sport = BPF_CORE_READ(ctx, sport);
	key.dport = BPF_CORE_READ(ctx, dport);
	account(&key);
	return 0;
}

SEC("tracepoint/tcp/tcp_retransmit_skb")
int argus_tcp_retransmit_skb(struct trace_event_raw_tcp_retransmit_skb *ctx)
{
	struct flow_key key = {};
	key.direction = FLOW_RETRANSMIT;

	__u16 family = BPF_CORE_READ(ctx, family);
	if (family == AF_INET) {
		key.family = 4;
		BPF_CORE_READ_INTO(&key.saddr, ctx, saddr);
		BPF_CORE_READ_INTO(&key.daddr, ctx, daddr);
	} else if (family == AF_INET6) {
		key.family = 6;
		BPF_CORE_READ_INTO(&key.saddr, ctx, saddr_v6);
		BPF_CORE_READ_INTO(&key.daddr, ctx, daddr_v6);
	} else {
		bump(STAT_SKIPPED);
		return 0;
	}

	key.sport = BPF_CORE_READ(ctx, sport);
	key.dport = BPF_CORE_READ(ctx, dport);
	account(&key);
	return 0;
}
