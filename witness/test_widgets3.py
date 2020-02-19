import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, Pango, Gdk, GLib
import datetime

class MainWin(Gtk.Window):
    width = 300
    height = int(width * 3/2)

    def __init__(self):
        super().__init__(border_width=0, title="ui test")
        self.set_size_request(MainWin.width, MainWin.height)

        lb1 = Gtk.ListBox()
        lb1.add(Gtk.Label("item 1"))
        lb1.add(Gtk.Label("item 2"))
        lb1.add(Gtk.Label("item 3"))

        lb2 = Gtk.ListBox()
        lb2.add(Gtk.Label("item 1"))
        lb2.add(Gtk.Label("item 2"))

        entry1 = Gtk.Entry()
        entry1.set_text("abc")

        def on_lb_focus_out(lb1, e):
            print("focus_out")
            lb.unselect_all()
        lb1.set_events(Gdk.EventMask.FOCUS_CHANGE_MASK)
        lb1.connect("focus-out-event", on_lb_focus_out)

#        entry1.set_events(Gdk.EventMask.FOCUS_CHANGE_MASK)
        entry1.connect("focus-out-event", on_lb_focus_out)

        vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=10)
        vbox.add(lb1)
        vbox.add(lb2)
        vbox.add(entry1)

        self.add(vbox)

        self.connect("destroy", Gtk.main_quit)
        self.show_all()


if __name__ == "__main__":
    w = MainWin()
    Gtk.main()

