import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk

def add_clicked(w, data):
  print("add_clicked: %s" % data)

def edit_clicked(w, data):
  print("edit_clicked: %s" % data)

w = Gtk.Window(border_width=10, title="Rob's Window")
w.connect("destroy", Gtk.main_quit)

#box = Gtk.Box(spacing=5, orientation=Gtk.Orientation.VERTICAL)

grid = Gtk.Grid()

bnAdd = Gtk.Button(label="Add")
bnAdd.connect("clicked", add_clicked, "123")

bnEdit = Gtk.Button(label="Edit")
bnEdit.connect("clicked", edit_clicked, "456")

#box.add(bnAdd)
#box.add(bnEdit)
#w.add(box)

#grid.attach(bnAdd, 1,0, 4,1)
grid.add(bnAdd)
grid.attach(bnEdit, 1,0, 2,1)
w.add(grid)

w.show_all()
Gtk.main()

