package common

// Prltype is a Python 2 script that sends scancodes to a Parallels VM. It requires
// the prlsdkapi Python module, which is bundled with the Parallels Virtualization SDK.
const Prltype string = `
import sys
import prlsdkapi


def main():
    if len(sys.argv) < 3:
        print "Usage: prltype VM_NAME SCANCODE..."
        sys.exit(1)

    vm_name = sys.argv[1]
    scancodes = sys.argv[2:]

    server = login()
    vm, vm_io = connect(server, vm_name)

    send(scancodes, vm, vm_io)

    disconnect(server, vm, vm_io)


def login():
    prlsdkapi.prlsdk.InitializeSDK(prlsdkapi.prlsdk.consts.PAM_DESKTOP_MAC)
    server = prlsdkapi.Server()
    login_job = server.login_local()
    login_job.wait()

    return server


def connect(server, vm_name):
    vm_list_job = server.get_vm_list()
    result = vm_list_job.wait()

    vm_list = [result.get_param_by_index(i) for i in range(result.get_params_count())]
    vm = [vm for vm in vm_list if vm.get_name() == vm_name]

    if not vm:
        vm_names = [vm.get_name() for vm in vm_list]
        raise Exception(
            "%s: No such VM. Available VM's are:\n%s" % (vm_name, "\n".join(vm_names))
        )

    vm = vm[0]

    vm_io = prlsdkapi.VmIO()
    vm_io.connect_to_vm(vm).wait()

    return (vm, vm_io)


def disconnect(server, vm, vm_io):
    if vm and vm_io:
        vm_io.disconnect_from_vm(vm)

    if server:
        server.logoff()

    prlsdkapi.deinit_sdk


def send(scancodes, vm, vm_io):
    delay = 100
    consts = prlsdkapi.prlsdk.consts

    for i, scancode in enumerate(scancodes):
        a = int(scancode, 16)
        if a == 224: # Scancode sequences start with e0
            b = int(scancodes.pop(i + 1), 16)
            if b < 128:
                vm_io.send_key_event(vm, (a, b), consts.PKE_PRESS, delay)
            else:
                vm_io.send_key_event(vm, (a, b - 128), consts.PKE_RELEASE, delay)
        elif a < 128:
            vm_io.send_key_event(vm, (a,), consts.PKE_PRESS, delay)
        else:
            vm_io.send_key_event(vm, (a - 128,), consts.PKE_RELEASE, delay)


if __name__ == "__main__":
    main()
`
