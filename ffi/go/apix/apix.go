package apix

// #cgo CFLAGS:
// #cgo LDFLAGS: -lapix
// #include <apix/apix.h>
// #include <apix/srrp.h>
// #include <apix/log.h>
import "C"
import "unsafe"
import "errors"
import "github.com/yonzkon/apix/ffi/go/srrp"

const (
    EventNone uint = 0
    EventOpen uint = 1
    EventClose uint = 2
    EventAccept uint = 3
    EventPollin uint = 4
)

type ApixStream struct {
    stream *C.struct_stream
    fd int
}

func (self *ApixStream) IsNull() (bool) {
    if self.stream == nil {
        return true
    } else {
        return false
    }
}

func (self *ApixStream) Close() (int) {
    return int(C.apix_close(self.stream))
}

func (self *ApixStream) Accept() (ApixStream) {
    new_stream := C.apix_accept(self.stream)
    return ApixStream {
        stream: new_stream,
            fd: int(C.apix_get_raw_fd(new_stream)),
    }
}

func (self *ApixStream) Send(buf []byte) (int) {
    return int(C.apix_send(self.stream,
        (*C.uchar)(unsafe.Pointer(&buf[0])), C.uint(len(buf))))
}

func (self *ApixStream) Recv(buf []byte) (int) {
    return int(C.apix_recv(self.stream,
        (*C.uchar)(unsafe.Pointer(&buf[0])), C.uint(len(buf))))
}

func (self *ApixStream) SendToBuffer(buf []byte) (int) {
    return int(C.apix_send_to_buffer(self.stream,
        (*C.uchar)(unsafe.Pointer(&buf[0])), C.uint(len(buf))))
}

func (self *ApixStream) ReadFromBuffer(buf []byte) (int) {
    return int(C.apix_read_from_buffer(self.stream,
        (*C.uchar)(unsafe.Pointer(&buf[0])), C.uint(len(buf))))
}

func (self *ApixStream) WaitEvent() (uint) {
    return uint(C.apix_wait_event(self.stream))
}

func from_raw_packet_apix(pac *C.struct_srrp_packet) (srrp.SrrpPacket) {
    if pac == nil {
        panic("raw point of srrp_packet is null")
    }

    return srrp.SrrpPacket {
        Leader: int8(C.srrp_get_leader(pac)),
        PacketLen: uint16(C.srrp_get_packet_len(pac)),
        Fin: uint8(C.srrp_get_fin(pac)),
        Ver: uint16(C.srrp_get_ver(pac)),
        PayloadLen: uint32(C.srrp_get_payload_len(pac)),
        Srcid: C.GoString(C.srrp_get_srcid(pac)),
        Dstid: C.GoString(C.srrp_get_dstid(pac)),
        Anchor: C.GoString(C.srrp_get_anchor(pac)),
        Payload: C.GoString((*C.char)(unsafe.Pointer(C.srrp_get_payload(pac)))),
        Crc16: uint16(C.srrp_get_crc16(pac)),
        Raw: C.GoBytes(unsafe.Pointer(C.srrp_get_raw(pac)),
            (C.int)(C.srrp_get_packet_len(pac))),
        Pac: unsafe.Pointer(pac),
    }
}

func (self *ApixStream) WaitSrrpPacket() (srrp.SrrpPacket, error) {
    pac := C.apix_wait_srrp_packet(self.stream)

    if pac == nil {
        return srrp.SrrpPacket {}, errors.New("wait srrp packet failed")
    } else {
        ret := from_raw_packet_apix(pac)
        return ret, nil
    }
}

func (self *ApixStream) UpgradeToSrrp(nodeid string) (int) {
    return int(C.apix_upgrade_to_srrp(self.stream, C.CString(nodeid)))
}

func (self *ApixStream) SrrpForward(pac srrp.SrrpPacket) () {
    C.apix_srrp_forward(self.stream, (*C.struct_srrp_packet)(pac.Pac))
}

func (self *ApixStream) SrrpSend(pac srrp.SrrpPacket) (int) {
    return int(C.apix_srrp_send(self.stream, (*C.struct_srrp_packet)(pac.Pac)))
}

type Apix struct {
    ctx *C.struct_apix
}

func New() (Apix) {
    return Apix{ ctx: C.apix_new() }
}

func (self *Apix) Drop() {
    C.apix_drop(self.ctx)
}

func (self *Apix) SetWaitTimeout(usec uint64) () {
    C.apix_set_wait_timeout(self.ctx, C.ulonglong(usec))
}

func (self *Apix) WaitStream() (ApixStream) {
    stream := C.apix_wait_stream(self.ctx)
    if stream == nil {
        return ApixStream{nil, -1}
    } else {
        return ApixStream{stream, int(C.apix_get_raw_fd(stream))}
    }
}

func (self *Apix) EnablePosix() {
    C.apix_enable_posix(self.ctx)
}

func (self *Apix) DisablePosix() {
    C.apix_disable_posix(self.ctx)
}

func (self *Apix) Open(sinkid string, addr string) (ApixStream) {
    stream := C.apix_open(self.ctx, C.CString(sinkid), C.CString(addr))
    return ApixStream {
        stream: stream,
            fd: int(C.apix_get_raw_fd(stream)),
    }
}

func (self *Apix) OpenUnixServer(addr string) (ApixStream) {
    return self.Open(C.SINK_UNIX_S, addr)
}

func (self *Apix) OpenUnixClient(addr string) (ApixStream) {
    return self.Open(C.SINK_UNIX_C, addr)
}

func (self *Apix) OpenTcpServer(addr string) (ApixStream) {
    return self.Open(C.SINK_TCP_S, addr)
}

func (self *Apix) OpenTcpClient(addr string) (ApixStream) {
    return self.Open(C.SINK_TCP_C, addr)
}

func (self *Apix) OpenCom(addr string) (ApixStream) {
    return self.Open(C.SINK_COM, addr)
}

func (self *Apix) OpenCan(addr string) (ApixStream) {
    return self.Open(C.SINK_CAN, addr)
}
