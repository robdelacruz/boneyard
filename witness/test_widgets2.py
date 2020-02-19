import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, Pango, Gdk, GLib
import datetime

import ui
import conv

def icon_image(icon_name):
    theme = Gtk.IconTheme.get_default()
    icon = theme.load_icon(icon_name, -1, Gtk.IconLookupFlags.FORCE_SIZE)
    img = Gtk.Image.new_from_pixbuf(icon)
    return img

class MainWin(Gtk.Window):
    width = 300
    height = int(width * 3/2)

    def __init__(self):
        super().__init__(border_width=0, title="ui test")
        self.set_size_request(MainWin.width, MainWin.height)

        txt = """Longer note. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party."""
        txt2 = """1. Longer note. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party.

2. Longer note. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party.

3. Longer note. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party. Now is the time for all good men to come to the aid of the party."""

        lbl = Gtk.Label()
        lbl.set_hexpand(True)
        lbl.set_xalign(0)
        lbl.set_yalign(0)
        lbl.set_line_wrap(True)
        lbl.set_ellipsize(Pango.EllipsizeMode.END)
        lbl.set_lines(2)
        lbl.set_max_width_chars(5)
        lbl.set_markup(txt2.strip().split("\n")[0])

        self.add(ui.frame(lbl, "heading"))

        self.connect("destroy", Gtk.main_quit)
        self.show_all()


if __name__ == "__main__":
    w = MainWin()
    Gtk.main()
