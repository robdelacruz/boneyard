import sys
import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, Pango, Gdk, GLib

import db
import ui
import conv
from widgets import JournalView, EditEntryWin

class Bag:
    pass


class MainWin(Gtk.Window):
#    width = 800
#    height = 600
    width = 1024
    height = int(width * 2/3)

    def __init__(self):
        super().__init__(border_width=0, title="ui test")
        self.set_size_request(MainWin.width, MainWin.height)

        con = None
        if len(sys.argv) > 1:
            dbfile = sys.argv[1]
            con = db.connect_db(dbfile)
        else:
            #con = db.connect_db()
            con = db.connect_db("test.db")

        db.init_tables(con)
        db.commit(con)
        self.con = con

        self.ui = Bag()
        self.state = Bag()

        #today_dt = conv.today_isodt()
        today_dt = "2019-06-04"
        self.state.dt = today_dt
        self.state.sel_entry = None

        self.setup_widgets()


    def setup_widgets(self):
        self.connect("destroy", Gtk.main_quit)

        # Content container
        vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=0)
        self.add(vbox)

        # Menu
        menubar = self.create_menubar()
        vbox.pack_start(menubar, False, False, 0)

        grid = Gtk.Grid()
        vbox.pack_start(grid, True, True, 0)

        col1 = self.create_col1()
        col2 = self.create_col2()
        col3 = self.create_col3()

        grid.attach(col1, 0,0, 1,1)
        grid.attach(col2, 1,0, 1,1)
        grid.attach(col3, 2,0, 1,1)

        self.show_all()


    def create_menubar(self):
        menubar = ui.create_menubar({
            "_Witness": ["_About", "_Quit"],
            "_Entry": ["New _Log", "New _Note", "New E_vent", "New _Task", "_Edit Selected", "_Delete Selected"],
            "_Search": ["_Entries",],
        }, self.on_activate)

        # accelerators
        CTRL = Gdk.ModifierType.CONTROL_MASK
        ALT = Gdk.ModifierType.MOD1_MASK
        SHIFT = Gdk.ModifierType.SHIFT_MASK

        accel_group = Gtk.AccelGroup.new()
        self.add_accel_group(accel_group)
        ui.set_menu_accelerators(menubar, accel_group, {
            "witness_quit":             [(ord('q'), CTRL)],
            "entry_new-log":            [(ord('l'), CTRL)],
            "entry_new-note":           [(ord('n'), CTRL)],
            "entry_new-event":          [(ord('v'), CTRL)],
            "entry_new-task":           [(ord('t'), CTRL)],
            "entry_edit-selected":      [(ord('e'), CTRL)],
            "entry_delete-selected":    [(ord('x'), CTRL)],
            "search_entries":           [(ord('k'), CTRL)],
        })

        return menubar


    def on_activate(self, w):
        tag = w.tag
        print(f"on_activate(): {tag}")

        if tag == "witness_quit":
            Gtk.main_quit()
        elif tag == "entry_new-log":
            self.start_edit_entry_modal(db.EntryType.LOG)
        elif tag == "entry_new-note":
            self.start_edit_entry_modal(db.EntryType.NOTE)
        elif tag == "entry_new-event":
            self.start_edit_entry_modal(db.EntryType.EVENT)
        elif tag == "entry_new-task":
            self.start_edit_entry_modal(db.EntryType.TASK)
        elif tag == "entry_edit-selected":
            if not self.state.sel_entry:
                return
            sel_entry_id = self.state.sel_entry.get("id")
            if not sel_entry_id:
                return
            editentry = EditEntryWin(self.con, dt=self.state.dt, on_save=self.on_save_entry, entry_id=sel_entry_id)
            editentry.set_modal(True)
            editentry.set_transient_for(self)
        elif tag == "entry_delete-selected":
            if not self.state.sel_entry:
                return
            sel_entry_id = self.state.sel_entry.get("id")
            if not sel_entry_id:
                return
            db.del_entry(self.con, sel_entry_id)
            db.commit(self.con)

            entry_type = self.state.sel_entry.get("type", db.EntryType.LOG)
            self.refresh_journalview(entry_type)


    def refresh_journalview(self, entry_type):
        if entry_type == db.EntryType.LOG:
            self.ui.logview.refresh()
        elif entry_type == db.EntryType.NOTE:
            self.ui.notesview.refresh()
        elif entry_type == db.EntryType.EVENT:
            self.ui.eventsview.refresh()
        elif entry_type == db.EntryType.TASK:
            self.ui.tasksview.refresh()


    def on_save_entry(self, entry):
        entry_type = entry.get("type", db.EntryType.LOG)
        self.refresh_journalview(entry_type)


    def start_edit_entry_modal(self, entrytype):
        editentry = EditEntryWin(self.con, dt=self.state.dt, on_save=self.on_save_entry, entrytype=entrytype)
        editentry.set_modal(True)
        editentry.set_transient_for(self)


    def create_col1(self):
        col1 = Gtk.Grid()
        col1.props.margin = 10
        col1.set_row_spacing(20)

        def on_sel_entry(entry):
            self.clear_sel_highlights(self.ui.logview)
            self.state.sel_entry = entry

        logview = JournalView(self.con, dt=self.state.dt, read_fn=db.read_today_logs, show_preview=True, on_sel_entry=on_sel_entry)

        self.ui.logview = logview
        col1.attach(logview.widget(), 0,0, 1,1)
        return col1


    def create_col2(self):
        col2 = Gtk.Grid()
        col2.props.margin = 10
        col2.set_row_spacing(20)
        col2.set_hexpand(False)

        cal = Gtk.Calendar()
        date = conv.isodt_to_date(self.state.dt)
        cal.props.year = date.year
        cal.props.month = date.month-1
        cal.props.day = date.day
        def on_sel_day(cal):
            (year, month, day) = cal.get_date()
            self.set_date(conv.dateparts_to_isodt(year, month+1, day))
        cal.connect("day-selected", on_sel_day)

        def on_events_sel_entry(entry):
            self.clear_sel_highlights(self.ui.eventsview)
            self.state.sel_entry = entry

        eventsview = JournalView(self.con, dt=self.state.dt, read_fn=db.read_today_events, heading="Events", on_sel_entry=on_events_sel_entry)

        def on_tasks_sel_entry(entry):
            self.clear_sel_highlights(self.ui.tasksview)
            self.state.sel_entry = entry

        tasksview = JournalView(self.con, dt=self.state.dt, read_fn=db.read_today_tasks, heading="Tasks", on_sel_entry=on_tasks_sel_entry)

        self.ui.cal = cal
        self.ui.eventsview = eventsview
        self.ui.tasksview = tasksview

        col2.attach(cal, 0,0, 1,1)
        col2.attach(eventsview.widget(), 0,1, 1,1)
        col2.attach(tasksview.widget(), 0,2, 1,1)
        return col2


    def create_col3(self):
        col = Gtk.Grid()
        col.props.margin = 10
        col.set_row_spacing(20)
        col.set_hexpand(False)
        col.set_size_request(300, -1)

        def on_sel_entry(entry):
            self.clear_sel_highlights(self.ui.notesview)
            self.state.sel_entry = entry

        notesview = JournalView(self.con, dt=self.state.dt, read_fn=db.read_recent_notes, heading="Recent Notes", show_preview=True, on_sel_entry=on_sel_entry)
        self.ui.notesview = notesview

        col.attach(notesview.widget(), 0,0, 1,1)
        return col


    def set_date(self, dt):
        self.state.dt = dt
        self.ui.logview.set_date(dt)
        self.ui.eventsview.set_date(dt)
        self.ui.tasksview.set_date(dt)




    def clear_sel_highlights(self, skiplb):
        for lb in [self.ui.logview, self.ui.eventsview, self.ui.tasksview, self.ui.notesview]:
            if lb == skiplb:
                continue
            lb.clear_sel_highlight()


### MainWin


if __name__ == "__main__":
    w = MainWin()
    Gtk.main()

