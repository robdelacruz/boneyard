import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk

def mktext_key_release(w, e, buf):
  print("mktext key release: %s" % e.string)

  s = buf.get_start_iter()
  e = buf.get_end_iter()
  print("buftxt: %s" % buf.get_text(s, e, True))

w = Gtk.Window(border_width=10, title="Rob's Window")
w.set_default_size(800, 600)
w.connect("destroy", Gtk.main_quit)

grid = Gtk.Grid()
grid.set_row_spacing(5)
grid.set_column_spacing(5)

mktext_tv = Gtk.TextView()
tvbuf = mktext_tv.get_buffer()
tvbuf.set_text("""text here""")
mktext_tv.connect("key-release-event", mktext_key_release, tvbuf)

mktext_sw = Gtk.ScrolledWindow()
mktext_sw.set_hexpand(True)
mktext_sw.set_vexpand(True)
mktext_sw.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC) 
mktext_sw.add(mktext_tv)

save = Gtk.Button(label="Save")
clear = Gtk.Button(label="Clear")

fields = Gtk.ListStore(str, str)
fields.append(["amt", "123.45"])
fields.append(["cat", "commute"])
fields.append(["footnote", "Here's where the story ends."])
fields.append(["field1", "Field 1."])
fields.append(["field2", "Field 2."])
fields.append(["field3", "Field 3."])
fields.append(["field4", "Field 4."])
fields.append(["field5", "Field 5."])
fields.append(["field6", "Field 6."])
fields.append(["field7", "Field 7."])
fields.append(["field8", "Field 8."])

fields_tv = Gtk.TreeView(fields)
#fields_tv.set_hexpand(False)
#fields_tv.set_vexpand(False)
ren = Gtk.CellRendererText()
col0 = Gtk.TreeViewColumn("field", ren, text=0)
col0.set_sort_column_id(0)
col1 = Gtk.TreeViewColumn("value", ren, text=1)
fields_tv.append_column(col0)
fields_tv.append_column(col1)

fields_sw = Gtk.ScrolledWindow()
fields_sw.set_hexpand(True)
fields_sw.set_vexpand(False)
fields_sw.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC) 
fields_sw.add(fields_tv)

grid.attach(mktext_sw, 0,0, 4,4)
grid.attach_next_to(save, mktext_sw, Gtk.PositionType.BOTTOM, 1,1)
grid.attach_next_to(clear, save, Gtk.PositionType.RIGHT, 1,1)

grid.attach_next_to(fields_sw, mktext_sw, Gtk.PositionType.RIGHT, 2,2)
w.add(grid)

w.show_all()
Gtk.main()

