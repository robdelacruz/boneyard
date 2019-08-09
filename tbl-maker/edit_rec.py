import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, GLib

import sqlite3

import util
import db

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


def rec_fields_preview(rec):
    # Get list of dot fields in alphabetical order.
    #   dot fields = [all rec fields] - [base fields]
    base_fields = {'_id', '_body'}
    fields = sorted(set(rec.keys()).difference(base_fields))

    # Get longest field name
    max_field_len = 0
    for field in fields:
        if len(field) > max_field_len:
            max_field_len = len(field)

    field_lines = []
    for field in fields:
        line = "<span weight='bold' font-family='monospace'>{k: <{w}}</span> : {v}".format(k=field, w=max_field_len, v=rec[field])
        field_lines.append(line)

    return "\n".join(field_lines)


EDITREC_WIDTH = 600
EDITREC_HEIGHT = int(EDITREC_WIDTH * 2/3)

class EditRecWin(Gtk.Window):
    def __init__(self, con, tbl, recid, parentw=None):
        super().__init__(border_width=10, title="Add Rec")
        self.set_default_size(EDITREC_WIDTH, EDITREC_HEIGHT)
        self.connect("destroy", self.on_destroy)

        self.con = con
        self.tbl = tbl
        self.parentw = parentw
        self.tbl_cols = db.table_cols(con, tbl)

        self.rec = None
        self.winid = ""

        # This is an 'Edit Rec' window if an existing recid was passed.
        # In this case, the recid will be used as the windows id.
        if recid:
            self.set_title("Edit Rec")
            self.rec = db.read_rec(con, tbl, recid)
            self.winid = recid

        # This is a 'New Rec' window if no recid passed or if nonexisting recid.
        # A window id will be programmatically generated.
        if self.rec == None:
            self.set_title("New Rec")
            self.rec = db.rec_new()
            self.winid = util.gen_id()

        # Grid
        self.grid = Gtk.Grid()
        self.add(self.grid)
        self.grid.set_row_spacing(5)
        self.grid.set_column_spacing(5)

        # Current Table display
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        lbl = Gtk.Label("Table: ")
        lbl_tbl = Gtk.Label()
        lbl_tbl.set_markup(f"<span weight='bold' font-family='monospace'>{self.tbl}</span>")
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
        sw_width = int(EDITREC_WIDTH * 1/2)
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

        # Save, Cancel buttons
        self.save = Gtk.Button(label="Save")
        self.cancel = Gtk.Button(label="Cancel")

        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        hbox.pack_start(self.save, False, False, 0)
        hbox.pack_start(self.cancel, False, False, 0)
        self.grid.attach(hbox, 0,6, 4,1)

        # Events
        self.mktext_tv.connect("key-release-event", self.mktext_key_release)
        self.save.connect("clicked", self.on_save)
        self.cancel.connect("clicked", self.on_cancel)

        self.show_all()
        self.refresh_ui()
        tv_move_cursor_start(self.mktext_tv)


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

        # Signal to parent that we're closing.
        if self.parentw:
            self.parentw.on_child_closed(self.winid)


    def on_save(self, w):
        buf = self.mktext_tv.get_buffer()
        s = buf.get_start_iter()
        e = buf.get_end_iter()
        mktext = buf.get_text(s, e, True)
        rec = db.rec_from_mktext(mktext)
        rec['_id'] = self.rec.get('_id')

        db.update_rec(self.con, self.tbl, rec)
        db.commit(self.con)

        # Signal to parent that rec was updated.
        if self.parentw:
            self.parentw.on_child_saved(rec)

        self.destroy()


    def on_cancel(self, w):
        self.destroy()


if __name__ == "__main__":
    con = db.connect_db()
    rec = db.rec_new()
    w = EditRecWin(con, "nbdata", "")
    Gtk.main()

