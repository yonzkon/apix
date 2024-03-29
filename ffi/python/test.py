import apix
import srrp

run_flag = 1

ctx = apix.Apix()
ctx.enable_posix()
ctx.set_wait_timeout(0)

stream = ctx.open_unix_client("/tmp/apix")
pac = srrp.srrp_new_ctrl("ff01", "/sync", "")
stream.send(pac.raw())

while run_flag:
    match stream.wait_event():
        case 1:
            print("open")
        case 2:
            print("close")
        case 3:
            stream.accept();
            print("never enter accept")
        case 4:
            data = stream.read_from_buffer()
            if len(data):
                pac = srrp.srrp_parse(data)
                if pac.is_null():
                    print(data)
                    if data.decode('utf-8') == 'exit':
                        run_flag = 0
                else:
                    print("recv srrp:", pac.raw())

stream.close()
