import threading

n_times = 0

def on_timer():
    global n_times
    print("on_timer()")

    n_times += 1
    if n_times >= 3:
        return

    t = threading.Timer(2.0, on_timer)
    t.start()

t = threading.Timer(2.0, on_timer)
t.start()

