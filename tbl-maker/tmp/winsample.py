import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk

class MainWin(Gtk.Window):
    def __init__(self):
        super().__init__(border_width=10, title="Main Win")
        self.set_default_size(500, 500)
        self.connect("destroy", Gtk.main_quit)

        # Grid
        self.grid = Gtk.Grid()
        self.add(self.grid)
        self.grid.set_row_spacing(5)
        self.grid.set_column_spacing(5)

        # Query, Clear buttons
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        self.search = Gtk.Button(label="Search")
        hbox.pack_start(self.search, False, False, 0)
        self.add = Gtk.Button(label="Add")
        hbox.pack_start(self.add, False, False, 0)
        self.grid.attach(hbox, 0,7, 10,1)

        # Events
        self.search.connect("clicked", self.on_search)
        self.add.connect("clicked", self.on_add)


    def on_search(self, w):
        self.child = Gtk.Window(border_width=10, title="Child Win")
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        self.child.add(hbox)

        ok = Gtk.Button(label="OK")
        hbox.pack_start(ok, False, False, 0)
        cancel = Gtk.Button(label="Cancel")
        hbox.pack_start(cancel, False, False, 0)

        self.child.show_all()


    def on_add(self, w):
        if self.child:
            Gtk.Widget.destroy(self.child)
        pass


def main():
    w = MainWin()
    w.show_all()
    Gtk.main()

if __name__ == "__main__":
    main()

