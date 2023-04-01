#ifndef __APIX_PRIVATE_H
#define __APIX_PRIVATE_H

#include <stdlib.h>
#include <time.h>
#include <sys/time.h>

#include "apix.h"
#include "types.h"
#include "list.h"
#include "vec.h"
#include "srrp.h"

#define SINK_ID_SIZE 64
#define STREAM_ADDR_SIZE 64

#define STREAM_SYNC_TIMEOUT (1000 * 5) /*ms*/
#define PARSE_PACKET_TIMEOUT 1000 /*ms*/
#define APIX_IDLE_MAX (1 * 1000 * 1000) /*us*/

#define PAYLOAD_LIMIT 1400

#ifdef __cplusplus
extern "C" {
#endif

/**
 * apix
 */

struct apix {
    struct list_head streams;
    struct list_head sinks;
    struct timeval poll_ts;
    u8 poll_cnt;
    u64 idle_usec;
};

/**
 * sink
 * - apix low level implement, maybe unix domain, bsd socket, uart, can ...
 */

struct sink;

struct sink_operations {
    int (*open)(struct sink *sink, const char *addr);
    int (*close)(struct sink *sink, int fd);
    int (*accept)(struct sink *sink, int fd);
    int (*ioctl)(struct sink *sink, int fd, unsigned int cmd, unsigned long arg);
    int (*send)(struct sink *sink, int fd, const u8 *buf, u32 len);
    int (*recv)(struct sink *sink, int fd, u8 *buf, u32 size);
    int (*poll)(struct sink *sink);
};

struct sink {
    char id[SINK_ID_SIZE]; // identify
    struct sink_operations ops;
    struct apix *ctx;
    struct list_head streams;
    struct list_head ln;
};

void sink_init(struct sink *sink, const char *id,
                  const struct sink_operations *ops);
void sink_fini(struct sink *sink);

int apix_sink_register(struct apix *ctx, struct sink *sink);
void apix_sink_unregister(struct apix *ctx, struct sink *sink);

/**
 * stream
 * - treat it as unix fd in most situations
 * - each sink holds several streams
 */

enum stream_type {
    STREAM_T_LISTEN = 'l',
    STREAM_T_ACCEPT = 'a',
    STREAM_T_CONNECT = 'c',
};

enum stream_state {
    STREAM_ST_NONE = 0,
    STREAM_ST_NODEID_NORMAL,
    STREAM_ST_NODEID_DUP,
    STREAM_ST_NODEID_ZERO,
    STREAM_ST_FINISHED,
};

struct stream {
    int fd;
    struct stream *father;
    char addr[STREAM_ADDR_SIZE];
    char type; /* stream_type */
    int state; /* stream_state */
    time_t ts_sync_in;
    time_t ts_sync_out;
    struct timeval ts_poll_recv;

    vec_8_t *txbuf;
    vec_8_t *rxbuf;

    union {
        u8 byte;
        struct {
            u8 open:1;
            u8 close:1;
            u8 accept:1;
            u8 pollin:1;
            u8 srrp_packet_in:1;
        } bits;
    } ev;

    // only for srrp
    int srrp_mode;
    u32 l_nodeid; /* local nodeid */
    u32 r_nodeid; /* remote nodeid */
    vec_p_t *sub_topics;
    struct srrp_packet *rxpac_unfin;
    struct list_head msgs;

    struct apix *ctx;
    struct sink *sink;
    struct list_head ln_ctx;
    struct list_head ln_sink;
};

struct stream *stream_new(struct sink *sink);
void stream_free(struct stream *stream);

struct stream *find_stream_in_apix(struct apix *ctx, int fd);
struct stream *find_stream_in_sink(struct sink *sink, int fd);
struct stream *find_stream_by_l_nodeid(struct apix *ctx, u32 nodeid);
struct stream *find_stream_by_r_nodeid(struct apix *ctx, u32 nodeid);
struct stream *find_stream_by_nodeid(struct apix *ctx, u32 nodeid);

/**
 * apimsg
 */

enum apimsg_state {
    APIMSG_ST_NONE = 0,
    APIMSG_ST_WAITING,
    APIMSG_ST_FINISHED,
    APIMSG_ST_FORWARD,
};

struct apimsg {
    int state;
    int fd; /* receive from fd */
    struct srrp_packet *pac;
    struct list_head ln;
};

static inline int apimsg_is_finished(struct apimsg *msg)
{
    return msg->state == APIMSG_ST_FINISHED;
}

static inline void apimsg_finish(struct apimsg *msg)
{
    msg->state = APIMSG_ST_FINISHED;
}

static inline void apimsg_free(struct apimsg *msg)
{
    list_del(&msg->ln);
    srrp_free(msg->pac);
    free(msg);
}

#ifdef __cplusplus
}
#endif
#endif
