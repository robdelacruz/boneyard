import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk

def save_clicked(w):
  print("save")

def clear_clicked(w):
  print("clear")


w = Gtk.Window(border_width=10, title="Rob's Window")
w.set_default_size(300, 300)
w.connect("destroy", Gtk.main_quit)

tv = Gtk.TextView()
tvbuf = tv.get_buffer()
tvbuf.set_text("""Now is the time
for all good men
to come to the aid
of the party.""")

sw = Gtk.ScrolledWindow()
sw.set_border_width(10)
sw.set_hexpand(True)
sw.set_vexpand(True)
sw.set_policy(Gtk.PolicyType.ALWAYS, Gtk.PolicyType.ALWAYS) 
#sw.add_with_viewport(tv)
sw.add(tv)

save = Gtk.Button(label="Save")
save.connect("clicked", save_clicked)

clear = Gtk.Button(label="Clear")
clear.connect("clicked", clear_clicked)

action_panel = Gtk.Box(spacing=5, orientation=Gtk.Orientation.HORIZONTAL)
action_panel.pack_start(save, False, False, 0)
action_panel.pack_start(clear, False, False, 0)

content = Gtk.Box(spacing=5, orientation=Gtk.Orientation.VERTICAL)
content.pack_start(sw, True, True, 0)
content.pack_start(action_panel, False, False, 0)

w.add(content)
w.show_all()
Gtk.main()

