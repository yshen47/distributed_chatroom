import selectors
import subprocess
import random
import sys

my_id = random.randrange(2**32)

print("{}: Starting log".format(my_id))
rate = float(sys.argv[1])

proc = subprocess.Popen(sys.argv[2:], stdin=subprocess.PIPE, stdout=subprocess.PIPE, bufsize=0)

sel = selectors.DefaultSelector()

# wait for READY, ignoring any lines with a hash mark
while True:
    l = proc.stdout.readline()
    if l.strip() == b"READY":
        print("{}: Process ready".format(my_id))
        break
    elif l[0] == '#':
        continue
    else:
        print("{}: Unexpected output: {}".format(my_id, l.strip()))
        sys.exit(1)

sel = selectors.DefaultSelector()
sel.register(proc.stdout, selectors.EVENT_READ)

msg_no = 0
event_no = 0
while True:
    event_no += 1
    # since it's memoryless we can always just resample
    # our delay
    events = sel.select(random.expovariate(rate))
    if not events:  # timeout
        msg_no += 1
        proc.stdin.write("{} message number {}\n".format(my_id, msg_no).encode())
        print("{}: {}: Sent message '{}'".format(my_id, event_no, msg_no))
    else:
        # FIXME check that events are what we expect
        l = proc.stdout.readline()
        if l[0] == '#':
            continue
        else:
            print("{}: {}: Received '{}'".format(my_id, event_no, l.decode().strip()))
