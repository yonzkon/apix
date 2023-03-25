package apix_test
import "github.com/yonzkon/apix/go/apix"
import "github.com/yonzkon/apix/go/srrp"
import "github.com/yonzkon/apix/go/log"
import "testing"
import "fmt"

func TestBase(T *testing.T) {
    log.LogSetLevelDebug()

    ctx := apix.New()
    ctx.EnablePosix()
    fd := ctx.OpenTcpClient("127.0.0.1:8080")

    pac, _ := srrp.NewCtrl(0x1111, "/sync", "")
    ctx.Send(fd, pac.Raw)

    for true {
        ctx.Poll(0)

        buf := make([]byte, 256)
        nr := ctx.ReadFromBuffer(fd, buf)
        if nr > 0 {
            pac, err := srrp.Parse(buf)
            if err == nil {
                fmt.Println("recv srrp: " + string(pac.Raw))
            } else {
                fmt.Println("recv raw: " + string(buf))
            }
        }
    }

    ctx.Close(fd)
    ctx.Destroy()
}
