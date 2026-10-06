/* Minimal kernel type declarations for the L4 collector.
 *
 * Only the tracepoint context fields actually read by l4.bpf.c are declared.
 * preserve_access_index makes clang emit CO-RE relocations, so field offsets
 * are resolved against the running kernel's BTF instead of being baked in.
 * This avoids vendoring a multi-megabyte vmlinux.h per architecture.
 */
#ifndef __ARGUS_KERNEL_DEFS_H__
#define __ARGUS_KERNEL_DEFS_H__

#pragma clang attribute push(__attribute__((preserve_access_index)), apply_to = record)

struct trace_event_raw_inet_sock_set_state {
	struct {
		char pad[8];
	} ent;
	const void *skaddr;
	int oldstate;
	int newstate;
	__u16 sport;
	__u16 dport;
	__u16 family;
	__u16 protocol;
	__u8 saddr[4];
	__u8 daddr[4];
	__u8 saddr_v6[16];
	__u8 daddr_v6[16];
};

struct trace_event_raw_tcp_retransmit_skb {
	struct {
		char pad[8];
	} ent;
	const void *skbaddr;
	const void *skaddr;
	int state;
	__u16 sport;
	__u16 dport;
	__u16 family;
	__u8 saddr[4];
	__u8 daddr[4];
	__u8 saddr_v6[16];
	__u8 daddr_v6[16];
};

#pragma clang attribute pop

#endif /* __ARGUS_KERNEL_DEFS_H__ */
