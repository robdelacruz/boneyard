import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, Pango, Gdk, GLib

import sqlite3

import util
import db
import ui

class EditRecWin(Gtk.Window):
    width = 800
    height = int(width * 2/3)

    def __init__(self, con, tbl, recid, on_edit_save=None, on_edit_close=None):
        super().__init__(border_width=0, title="")
        self.set_size_request(EditRecWin.width, EditRecWin.height)

        self.setup(con, tbl, recid, on_edit_save, on_edit_close)


    def setup(self, con, tbl, recid, on_edit_save=None, on_edit_close=None):
        self.connect("destroy", self.on_destroy)

        self.con = con
        self.tbl = tbl
        self.on_edit_save = on_edit_save
        self.on_edit_close = on_edit_close
        self.tbl_cols = db.table_cols(con, tbl)

        self.rec = None
        self.winid = ""
        self.is_new_rec = False

        # This is an 'Edit Rec' window if an existing recid was passed.
        # In this case, the recid will be used as the windows id.
        if recid:
            self.set_title("Edit Record")
            self.rec = db.read_rec(con, tbl, recid)
            self.winid = recid

        # This is a 'New Rec' window if no recid passed or if nonexisting recid.
        # A window id will be programmatically generated.
        if self.rec == None:
            self.set_title("New Record")
            self.rec = db.rec_new()
            self.winid = util.gen_id()
            self.is_new_rec = True

        self.setup_widgets()


    def setup_widgets(self):
        # accelerators
        CTRL = Gdk.ModifierType.CONTROL_MASK
        ALT = Gdk.ModifierType.MOD1_MASK
        SHIFT = Gdk.ModifierType.SHIFT_MASK

        widget_accel_group = Gtk.AccelGroup.new()
        self.add_accel_group(widget_accel_group)

        # Content container
        vbox = Gtk.Box(orientation=Gtk.Orientation.VERTICAL, spacing=0)
        self.add(vbox)

        # Menu
        menubar = self.create_menubar()
        vbox.pack_start(menubar, False, False, 0)

        # Grid
        self.grid = Gtk.Grid()
        self.grid.props.margin = 10
        self.grid.set_row_spacing(5)
        self.grid.set_column_spacing(5)
        vbox.pack_start(self.grid, True, True, 0)

        # Current Table display
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        lbl = Gtk.Label("Table: ")
        lbl_tbl = Gtk.Label()
        lbl_tbl.set_markup(f"<span weight='bold' font-family='monospace'>{ui.escape(self.tbl)}</span>")
        hbox.pack_start(lbl, False, False, 0)
        hbox.pack_start(lbl_tbl, False, False, 0)
        self.grid.attach(hbox, 0,0, 10,1)

        # Horizontal box containing textview and fields
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        self.grid.attach(hbox, 0,1, 8,4)

        # Textview in scrolledwindow
        self.mktext_sw = Gtk.ScrolledWindow()
        hbox.pack_start(self.mktext_sw, True, True, 0)
        self.mktext_sw.set_size_request(300, 0)

        self.mktext_sw.set_hexpand(True)
        self.mktext_sw.set_vexpand(True)
        self.mktext_sw.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC) 

        self.mktext_tv = Gtk.TextView()
        self.mktext_sw.add(self.mktext_tv)
        self.mktext_tv.set_wrap_mode(Gtk.WrapMode.WORD)

        # Set textview width
        sw_width = int(EditRecWin.width * 1/2)
        self.mktext_sw.set_size_request(sw_width, 0)

        # Fields listbox
        # show_fields() to show fields, hide_fields() to hide.
        # fields listbox is initially hidden until dotfields are entered or parsed.
        self.fields_sw = Gtk.ScrolledWindow()
        hbox.pack_start(self.fields_sw, True, True, 0)
        self.fields_sw.set_hexpand(True)
        self.fields_sw.set_vexpand(False)
        self.fields_sw.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC) 

        self.fields_lb = Gtk.ListBox()
        self.fields_sw.add(self.fields_lb)
        self.fields_lb.set_selection_mode(Gtk.SelectionMode.NONE)

        self.hide_fields()

        # Action buttons
        close = Gtk.Button(label="Close")
        close.tag = "record_close"
        close.connect("clicked", self.on_activate)

        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        hbox.pack_start(close, False, False, 0)
        self.grid.attach(hbox, 0,6, 4,1)

        # Events
        self.mktext_tv.connect("key-release-event", self.mktext_key_release)

        self.show_all()
        self.refresh_ui()
        tv_move_cursor_start(self.mktext_tv)


    def create_menubar(self):
        if self.is_new_rec:
            items = ["_Save", "_Discard", "_Close"]
        else:
            items = ["_Save", "_Close"]

        menubar = ui.create_menubar({
            "_Record": items,
        }, self.on_activate)

        # accelerators
        CTRL = Gdk.ModifierType.CONTROL_MASK
        ALT = Gdk.ModifierType.MOD1_MASK
        SHIFT = Gdk.ModifierType.SHIFT_MASK

        accel_group = Gtk.AccelGroup.new()
        self.add_accel_group(accel_group)

        if self.is_new_rec:
            # In New Rec mode, ESC discards any edits and closes window.
            ui.set_menu_accelerators(menubar, accel_group, {
                "record_save":      [(ord('s'), CTRL)],
                "record_discard":   [(Gdk.KEY_Escape, 0)],
                "record_close":     [(ord('w'), CTRL)],
            })
        else:
            # In Edit Rec mode, ESC and CTRL-W both save and close the window.
            ui.set_menu_accelerators(menubar, accel_group, {
                "record_save":      [(ord('s'), CTRL)],
                "record_close":     [(ord('w'), CTRL), (Gdk.KEY_Escape, 0)],
            })

        return menubar


    def refresh_ui(self):
        mktext = db.mktext_from_rec(self.rec)
        self.mktext_tv.get_buffer().set_text(mktext)
        self.refresh_fields(mktext)


    def refresh_fields(self, mktext):
        rec = db.rec_from_mktext(mktext)

        # Get list of dot fields in alphabetical order.
        #   dot fields = [all rec fields] - [base fields]
        base_fields = {'_id', '_body'}
        fields = sorted(set(rec.keys()).difference(base_fields))

        # Add undefined table cols at the end of fields list.
        for col in sorted(self.tbl_cols):
            if not col in fields:
                fields.append(col)

        # Get longest field name
        max_field_len = 0
        for field in fields:
            if field.startswith("_"):
                continue
            if len(field) > max_field_len:
                max_field_len = len(field)

        clear_lb(self.fields_lb)
        for field in fields:
            if field.startswith("_"):
                continue

            row = Gtk.ListBoxRow()
            lbl = Gtk.Label()

            # Show field definition as listbox row label.
            k = "{k: <{w}}".format(k=field, w=max_field_len)

            v = ""
            if field in rec:
                v = escape(str(rec[field]))

            if field in self.tbl_cols:
                lbl.set_markup(f"<span weight='bold' font-family='monospace'>{k}</span> : {v}")
            else:
                lbl.set_markup(f"<span foreground='green' weight='bold' font-family='monospace'>{k}</span> : {v}")

            lbl.set_xalign(0)
            lbl.set_line_wrap(True)
            row.add(lbl)
            self.fields_lb.add(row)

        self.fields_lb.show_all()

        # Show fields section only if there are any.
        if len(fields) > 0:
            self.show_fields()
        else:
            self.hide_fields()


    def show_fields(self):
        self.fields_sw.show()


    def hide_fields(self):
        self.fields_sw.hide()


    def mktext_key_release(self, w, data):
        mktext = tv_get_text(self.mktext_tv)
        self.refresh_fields(mktext)


    def on_destroy(self, w):
        if __name__ == "__main__":
            Gtk.main_quit()

        if self.on_edit_close:
            self.on_edit_close(self.winid)


    def run_save(self):
        buf = self.mktext_tv.get_buffer()
        s = buf.get_start_iter()
        e = buf.get_end_iter()
        mktext = buf.get_text(s, e, True)

        # Don't save rec if there's no content.
        if self.is_new_rec and mktext.strip() == "":
            return

        rec = db.rec_from_mktext(mktext)
        rec['_id'] = self.rec.get('_id')

        db.update_rec(self.con, self.tbl, rec)
        db.commit(self.con)

        self.rec = rec

        if self.on_edit_save:
            self.on_edit_save(rec, self.tbl)


    def on_activate(self, w):
        tag = w.tag
        print(f"on_activate(): {tag}")

        if tag == "record_save":
            self.run_save()
        elif tag == "record_discard":
            self.destroy()
        elif tag == "record_close":
            self.run_save()
            self.destroy()


# Helper functions
def escape(s):
    return GLib.markup_escape_text(s)


def clear_lb(lb):
    lb.foreach(lambda row: lb.remove(row))


def tv_get_text(tv):
    buf = tv.get_buffer()
    s = buf.get_start_iter()
    e = buf.get_end_iter()
    return buf.get_text(s, e, True)


def tv_move_cursor_start(tv):
    buf = tv.get_buffer()
    s = buf.get_start_iter()
    buf.place_cursor(s)


if __name__ == "__main__":
    con = db.connect_db()
    rec = db.rec_new()
    w = EditRecWin(con, "nbdata", "")
    Gtk.main()

