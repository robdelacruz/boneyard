import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk, Pango, Gdk, GLib

import sqlite3

import db
from edit_rec import EditRecWin

def escape(s):
    return GLib.markup_escape_text(s)


def rec_sorted_fields(rec):
    ks = set(rec.keys())
    ks.discard('_id')
    ks.discard('_body')
    return sorted(ks)


def rec_row_excerpt(rec):
    body = rec.get('_body') or ""
    body_lines = body.split("\n")

    # Return first line in body that has content. Skip over blank lines.
    for line in body_lines:
        if line.strip() != "":
            return line

    return "(empty content)"


def rec_preview(rec):
    body = rec.get('_body') or ""
    body = escape(body)
    return body + "\n" + rec_fields_preview(rec)


def rec_fields_preview(rec):
    # Get list of dot fields in alphabetical order.
    #   dot fields = [all rec fields] - [base fields]
    base_fields = {'_id', '_body'}
    fields = sorted(set(rec.keys()).difference(base_fields))

    # Get longest field name
    max_field_len = 0
    for field in fields:
        if not db.is_valid_field_name(field):
            continue
        if len(field) > max_field_len:
            max_field_len = len(field)

    field_lines = []
    for field in fields:
        if not db.is_valid_field_name(field):
            continue

        k = "{k: <{w}}".format(k=field, w=max_field_len)
        v = escape(str(rec[field]))

        line = f"<span weight='bold' font-family='monospace'>{k}</span> : {v}"
        field_lines.append(line)

    return "\n".join(field_lines)


def add_header_sep(w, data):
    sep = Gtk.Separator(orientation=Gtk.Orientation.HORIZONTAL)
    w.set_header(sep)


class CreateTableDlg(Gtk.Dialog):
    def __init__(self, parent):
        Gtk.Dialog.__init__(self, "Create Table", parent, Gtk.DialogFlags.MODAL,
            (Gtk.STOCK_OK, Gtk.ResponseType.OK,
            Gtk.STOCK_CANCEL, Gtk.ResponseType.CANCEL))
        self.set_border_width(10)

        content = self.get_content_area()
        content.set_spacing(10)
        lbl = Gtk.Label("New table:")
        lbl.set_xalign(0)
        content.add(lbl)

        self.tbl_entry = Gtk.Entry()
        self.tbl_entry.set_width_chars(20)
        self.tbl_entry.grab_focus()
        content.add(self.tbl_entry)

        self.show_all()


    def table_name(self):
        tbl = self.tbl_entry.get_text().strip()
        return tbl



QUERYREC_WIDTH = 1000
QUERYREC_HEIGHT = int(QUERYREC_WIDTH * 2/3)

class QueryWin(Gtk.Window):
    def __init__(self):
        super().__init__(border_width=10, title="Query Rec")
        self.set_default_size(QUERYREC_WIDTH, QUERYREC_HEIGHT)
#        self.modify_bg(Gtk.StateType.NORMAL, Gdk.Color(0xDDDD, 0xDDDD, 0xDDDD))

        self.connect("destroy", Gtk.main_quit)
        self.connect("key-press-event", self.on_keypress)

        self.con = db.connect_db()

        # 'New' and 'Edit' child windows
        # indexed by a generated id.
        self.child_wins = {}

        # Remember currently selected rec.
        self.sel_rec = None

        # Remember last query parms 
        self.tbl = ""
        self.sfind = ""
        self.qwhere = ""

        # Grid
        self.grid = Gtk.Grid()
        self.add(self.grid)
        self.grid.set_row_spacing(5)
        self.grid.set_column_spacing(5)

        # Table combotext select
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=10)
        self.grid.attach(hbox, 0,0, 5,1)
        lbl = Gtk.Label("Select Table ")
        self.tblsel = Gtk.ComboBoxText()
        hbox.pack_start(self.tblsel, False, False, 0)
        hbox.pack_start(lbl, False, False, 0)

        self.refresh_tables("nbdata")

        # Create Table action button
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        self.grid.attach(hbox, 6,0, 5,1)
        create_table = Gtk.Button(label="Create Table")
        hbox.pack_end(create_table, False, False, 0)

        # Find, Where, Order By entry
        lbl = Gtk.Label("Find in text")
        lbl.set_xalign(0)
        self.grid.attach(lbl, 0,1, 10,1)

        self.find_entry = Gtk.Entry()
        self.find_entry.set_width_chars(100)
        self.grid.attach(self.find_entry, 0,2, 10,1)

        lbl = Gtk.Label("Where query")
        lbl.set_xalign(0)
        self.grid.attach(lbl, 0,3, 10,1)

        self.where_entry = Gtk.Entry()
        self.where_entry.set_width_chars(100)
        self.grid.attach(self.where_entry, 0,4, 10,1)

        # Query, Clear buttons
        self.query = Gtk.Button(label="Query")
        self.clear = Gtk.Button(label="Clear")

        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        hbox.pack_start(self.query, False, False, 0)
        hbox.pack_start(self.clear, False, False, 0)
        self.grid.attach(hbox, 0,5, 10,1)

        self.hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=10)
        self.grid.attach(self.hbox, 0,6, 10,4)

        # Results listbox
        self.results_sw = Gtk.ScrolledWindow()
        self.hbox.pack_start(self.results_sw, True, True, 0)
        self.results_sw.set_hexpand(True)
        self.results_sw.set_vexpand(True)
        self.results_sw.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC)

        sw_width = int(QUERYREC_WIDTH * 1/3)
        self.results_sw.set_size_request(sw_width, 0) # set results width

        self.add_results_listbox()

        # Rec Preview multiline label
        sw = Gtk.ScrolledWindow()
        self.hbox.pack_start(sw, True, True, 0)
        sw.set_hexpand(True)
        sw.set_vexpand(True)
        sw.set_policy(Gtk.PolicyType.AUTOMATIC, Gtk.PolicyType.AUTOMATIC)

        frame = Gtk.Frame()
        sw.add(frame)

        self.preview = Gtk.Label()
        frame.add(self.preview)
        self.preview.set_xalign(0)
        self.preview.set_yalign(0)
        self.preview.set_line_wrap(True)

        # New, Edit, Del action buttons
        hbox = Gtk.Box(orientation=Gtk.Orientation.HORIZONTAL, spacing=5)
        self.grid.attach(hbox, 7,10, 3,1)

        self.delete = Gtk.Button(label="Del")
        hbox.pack_end(self.delete, False, False, 0)

        self.edit = Gtk.Button(label="Edit")
        hbox.pack_end(self.edit, False, False, 0)

        self.new = Gtk.Button(label="New")
        hbox.pack_end(self.new, False, False, 0)

        # No rec selected initially, so disable Edit/Del buttons.
        self.edit.set_sensitive(False)
        self.delete.set_sensitive(False)

        # Events
        self.tblsel.connect("changed", self.on_tblsel_changed)
        create_table.connect("clicked", self.on_create_table)

        self.query.connect("clicked", self.on_query)
        self.clear.connect("clicked", self.on_clear)

        self.new.connect("clicked", self.on_new)
        self.edit.connect("clicked", self.on_edit)
        self.delete.connect("clicked", self.on_delete)

        self.find_entry.grab_focus()


    def add_results_listbox(self):
        self.results = Gtk.ListBox()
        self.results_sw.add(self.results)
        self.results.set_header_func(add_header_sep)
        self.results.set_activate_on_single_click(False) # Need double click for row activate
        self.results.set_can_focus(True)

        self.results.connect("row-selected", self.on_results_select)
        self.results.connect("row-activated", self.on_results_activate)

        self.results_sw.show_all()


    def clear_results(self):
        self.results_sw.remove(self.results)
        self.add_results_listbox()

        # Clear preview pane.
        self.preview.set_markup("")


    def refresh_tables(self, active_tbl):
        """ Refresh table comboboxtext and initialize selection to active_tbl. """

        # Get tables list sorted alphabetically with 'nbdata' appearing first.
        all_tbls = list(db.all_tables(self.con))
        if "nbdata" in all_tbls:
            all_tbls.remove("nbdata")
        all_tbls.sort()
        all_tbls.insert(0, "nbdata")

        active_tbl_idx = 0
        self.tblsel.remove_all()
        for tbl in all_tbls:
            self.tblsel.append_text(tbl)

            if active_tbl == tbl:
                self.tblsel.set_active(active_tbl_idx)

            active_tbl_idx += 1


    def on_keypress(self, w, e):
        keyname = Gdk.keyval_name(e.keyval)

        is_ctrl = e.state & Gdk.ModifierType.CONTROL_MASK

        # CTRL-K
        if is_ctrl and keyname == 'k':
            self.find_entry.grab_focus()

        if keyname == 'F5':
            self.query.emit("clicked")


    def on_tblsel_changed(self, w):
        self.tbl = self.tblsel.get_active_text()
        self.clear_all()


    def on_query(self, w):
        tbl = self.tblsel.get_active_text()
        sfind = self.find_entry.get_text().strip()
        qwhere = self.where_entry.get_text().strip()

        try:
            self.run_query(tbl, sfind, qwhere)
        except Exception as e:
            dlg = Gtk.MessageDialog(self, Gtk.DialogFlags.MODAL, Gtk.MessageType.INFO,
                Gtk.ButtonsType.OK, "SQL Error")
            dlg.format_secondary_text(str(e))
            dlg.run()

            dlg.destroy()
            return

        self.results.grab_focus()

        # Remember last successful query parms 
        self.tbl = tbl
        self.sfind = sfind
        self.qwhere = qwhere

        # No rec selected initially, so disable Edit/Del buttons.
        self.edit.set_sensitive(False)
        self.delete.set_sensitive(False)


    def run_query(self, tbl, sfind, qwhere):
        recs = db.read_recs(self.con, tbl, sfind, qwhere)

        self.clear_results()
        for rec in recs:
            row = Gtk.ListBoxRow()
            row.rec = rec

            body = escape(rec_row_excerpt(rec))
            lbl = Gtk.Label()
            lbl.set_xalign(0)
            lbl.set_line_wrap(True)
            lbl.set_ellipsize(Pango.EllipsizeMode.END)
            lbl.set_lines(3)
            lbl.set_markup(body)
            row.add(lbl)

            self.results.add(row)

        self.results.show_all()


    def on_clear(self, w):
        self.clear_all()


    def clear_all(self):
        # Clear UI
        self.find_entry.set_text("")
        self.where_entry.set_text("")
        self.find_entry.grab_focus()
        self.clear_results()


    def on_results_select(self, w, row):
        if row:
            # Remember currently selected rec.
            self.sel_rec = row.rec

            # Update preview pane.
            self.preview.set_markup(rec_preview(row.rec))

            # A rec was selected, so enable Edit/Del buttons.
            self.edit.set_sensitive(True)
            self.delete.set_sensitive(True)
        else:
            # Clear preview pane.
            self.preview.set_markup("")

            # No rec selected, so disable Edit/Del buttons.
            self.edit.set_sensitive(False)
            self.delete.set_sensitive(False)


    # Results listbox double-click is same as selecting row and clicking 'Edit' button.
    def on_results_activate(self, w, row):
        self.on_edit(w)


    def on_create_table(self, w):
        dlg = CreateTableDlg(self)
        resp = dlg.run()

        # Create new table.
        if resp == Gtk.ResponseType.OK:
            tbl = dlg.table_name()
            if tbl != "":
                db.init_table(self.con, tbl)
                db.commit(self.con)
                self.refresh_tables(tbl)

        dlg.destroy()


    def on_new(self, w):
        tbl = self.tblsel.get_active_text()

        # Show Edit Rec window with blank recid.
        childw = EditRecWin(self.con, tbl, "", self)
        self.child_wins[childw.winid] = childw


    def on_edit(self, w):
        tbl = self.tblsel.get_active_text()
        row = self.results.get_selected_row()
        rec = row.rec
        recid = rec.get('_id', "")

        # If rec is currently being edited in a window, show that window.
        # Else pop up a new edit window for that rec.
        if recid in self.child_wins:
            self.child_wins[recid].present()
        else:
            childw = EditRecWin(self.con, tbl, recid, self)
            self.child_wins[childw.winid] = childw


    def on_delete(self, w):
        tbl = self.tblsel.get_active_text()
        row = self.results.get_selected_row()
        rec = row.rec
        recid = rec.get('_id', "")

        dlg = Gtk.MessageDialog(self, 0, Gtk.MessageType.QUESTION, Gtk.ButtonsType.YES_NO, "Delete this rec?")
        dlg.format_secondary_text(rec_row_excerpt(rec))
        resp = dlg.run()
        dlg.destroy()

        if resp == Gtk.ResponseType.YES:
            # Delete rec
            db.del_rec(self.con, tbl, recid)
            db.commit(self.con)

            # Close any open edit window for this rec.
            if recid in self.child_wins:
                childw = self.child_wins[recid]
                childw.destroy()
                self.child_wins.pop(recid, None)

            # Refresh result listbox to reflect changes.
            self.run_query(self.tbl, self.sfind, self.qwhere)
            self.restore_results_selection()


    def restore_results_selection(self):
        if self.sel_rec == None:
            return

        # Restore last rec selected from on_results_select().
        for row in self.results:
            if row.rec['_id'] == self.sel_rec['_id']:
                self.results.select_row(row)
                break


    def on_child_closed(self, winid):
        # Child window has signaled parent window that it closed.
        # Remove child window from child_wins dict.
        self.child_wins.pop(winid, None)


    def on_child_saved(self, rec):
        """Called when a rec is updated in a child window."""

        # Get currently selected row rec
        sel_recid = ""
        sel_row = self.results.get_selected_row()
        if sel_row:
            sel_recid = sel_row.rec['_id']

        for row in self.results:
            # Skip over rows not matching the rec.
            if row.rec['_id'] != rec['_id']:
                continue

            # Update row with new rec text.
            row.rec = rec
            for lbl in row:
                body = escape(rec_row_excerpt(rec))
                lbl.set_markup(body)
                break

            row.show_all()
            break
        else:
            # A new rec was added. Refresh result listbox to reflect changes.
            self.run_query(self.tbl, self.sfind, self.qwhere)
            self.restore_results_selection()


        # If row rec is selected, update preview pane.
        if sel_recid == rec['_id']:
            self.preview.set_markup(rec_preview(rec))


if __name__ == "__main__":
    w = QueryWin()
    w.show_all()
    Gtk.main()

