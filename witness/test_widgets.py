import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, Pango, Gdk, GLib
import datetime

import ui
import conv

class MainWin(Gtk.Window):
    width = 300
    height = int(width * 3/2)

    def __init__(self):
        super().__init__(border_width=0, title="ui test")
        self.set_size_request(MainWin.width, MainWin.height)

        grid1 = Gtk.Grid()
        lbl = Gtk.Label("Date Entry")
        de1 = DateEntry()
        de2 = DateEntry()
        de2.set_isodt("2019-01-02")
        grid1.attach(lbl, 0,0, 1,1)
        grid1.attach(de1.widget(), 0,1, 1,1)
        grid1.attach(de2.widget(), 0,2, 1,1)

        grid2 = Gtk.Grid()
        lbl = Gtk.Label("Form 2")
        chk = Gtk.CheckButton("Check 1")
        grid2.attach(lbl, 0,0, 1,1)
        grid2.attach(chk, 0,1, 1,1)

        grid3 = Gtk.Grid()
        lbl = Gtk.Label("Form 3")
        grid3.attach(lbl, 0,0, 1,2)
        grid3.set_hexpand(True)

        stack = Gtk.Stack()
        stack.add_titled(grid1, "pane1", "Journal")
        stack.add_titled(grid2, "pane2", "Topics")
        stack.add_titled(grid3, "pane3", "Utility")

        ss = Gtk.StackSwitcher()
        ss.set_stack(stack)
        ss.set_halign(Gtk.Align.CENTER)

        grid = Gtk.Grid()
        grid.attach(ss, 0,0, 1,1)
        grid.attach(ui.frame(stack), 0,1, 1,1)
        self.add(grid)

        self.connect("destroy", Gtk.main_quit)
        self.show_all()


class DateEntry():
    def __init__(self):
        entry = Gtk.Entry()
        entry.set_icon_from_icon_name(Gtk.EntryIconPosition.SECONDARY, "x-office-calendar")

        popover = Gtk.Popover()
        cal = Gtk.Calendar()
        popover.add(cal)
        popover.set_position(Gtk.PositionType.BOTTOM)
        popover.set_relative_to(entry)

        def on_icon_clicked(entry, *args):
            date = conv.isodt_to_date(entry.get_text())
            if not date:
                date = datetime.datetime.now()

            cal.props.year = date.year
            cal.props.month = date.month-1
            cal.props.day = date.day

            popover.show_all()
            popover.popup()
        entry.connect("icon-press", on_icon_clicked)

        def on_sel_day(cal):
            if cal.is_visible():
                (year, month, day) = cal.get_date()
                month += 1
                entry.set_text(conv.dateparts_to_isodt(year, month, day))
        cal.connect("day-selected", on_sel_day)

        def on_sel_day_dblclick(cal):
            popover.popdown()
        cal.connect("day-selected-double-click", on_sel_day_dblclick)

        self.entry = entry


    def widget(self):
        return self.entry

    def set_date(self, date):
        self.entry.set_text(conv.date_to_isodt())

    def get_date(self):
        date = conv.isodt_to_date(self.entry.get_text())
        if not date:
            date = datetime.datetime.now()
        return date

    def set_isodt(self, isodt):
        self.entry.set_text(isodt)

    def get_isodt(self):
        return self.entry.get_text()


if __name__ == "__main__":
    w = MainWin()
    Gtk.main()
